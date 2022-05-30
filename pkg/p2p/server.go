package p2p

import (
	"crypto/sha256"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sunboyy/lettered/pkg/security"
)

// HandlerFunc defines the handler used by P2P service.
type HandlerFunc func(publicKey string, body []byte) (interface{}, error)

// GinHandler creates a gin handler wrapping P2P handler function. It validates
// the request and extracts the request information before sending the request
// to the wrapped handler function. If the public key or the signature is
// invalid, the error response will be returned immediately.
func GinHandler(handler HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		publicKeyString := ctx.Request.Header.Get(headerPublicKey)
		signatureString := ctx.Request.Header.Get(headerSignature)

		publicKey, err := security.ParsePublicKey(publicKeyString)
		if err != nil {
			respondError(
				ctx,
				http.StatusUnauthorized,
				err,
				"invalid public key",
			)
			return
		}

		body, err := ioutil.ReadAll(ctx.Request.Body)
		if err != nil {
			respondError(
				ctx,
				http.StatusBadRequest,
				err,
				"cannot read request body",
			)
			return
		}

		// Calculate SHA256 of timestamp header for verifying with the
		// signature.
		hash := sha256.Sum256(body)

		if !security.VerifySignature(
			publicKey,
			hash[:],
			signatureString,
		) {
			ctx.JSON(
				http.StatusUnauthorized,
				gin.H{"error": "invalid signature"},
			)
			return
		}

		response, err := handler(publicKeyString, body)
		if err != nil {
			ctx.JSON(
				http.StatusBadRequest,
				gin.H{"error": err.Error()},
			)
			return
		}

		ctx.JSON(http.StatusOK, response)
	}
}

func respondError(ctx *gin.Context, code int, err error, message string) {
	log.Debug().Str("source", "p2p.HandlerWrapper").Err(err).Msg(message)
	ctx.JSON(code, gin.H{"error": message})
}
