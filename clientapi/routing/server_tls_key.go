package routing

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/matrix-org/util"

	"github.com/element-hq/dendrite/setup/config"
	"github.com/matrix-org/gomatrixserverlib/fclient"
	"github.com/matrix-org/gomatrixserverlib/spec"
)

// the response returned to the client
type serverTLSCertResponse struct {
	ServerName        string `json:"server_name"`
	PublicKeyBase64   string `json:"public_key_base64,omitempty"`
	ValidationStatus  string `json:"validation_status"`
	Error             string `json:"error,omitempty"`
    FingerprintSHA256 string `json:"fingerprint_sha256,omitempty"` 
}

func QueryServerTLSCertificate(ctx context.Context, targetServerNameStr string, cfg *config.ClientAPI, federation fclient.FederationClient) util.JSONResponse {
	logger := util.GetLogger(ctx).WithField("target_server", targetServerNameStr)
	targetServerName := spec.ServerName(targetServerNameStr)

	resp := serverTLSCertResponse{
		ServerName: targetServerNameStr,
	}

	cert, err := federation.GetServerTLSCertificate(ctx, targetServerName)
	if err != nil {
		resp.ValidationStatus = "unreachable_or_untrusted"
		resp.Error = err.Error()
		logger.WithError(err).Warn("Failed to get or validate TLS certificate from federationapi")
		return util.JSONResponse{
			Code: http.StatusInternalServerError, // Or more specific errors like 502 for upstream issues
			JSON: resp,
		}
	}
	
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(cert.PublicKey)
	if err != nil {
		resp.ValidationStatus = "error_processing_key"
		resp.Error = fmt.Sprintf("Failed to marshal public key: %v", err)
		logger.WithError(err).Error("Failed to marshal public key from certificate")
		return util.JSONResponse{
			Code: http.StatusInternalServerError,
			JSON: resp,
		}
	}

	resp.PublicKeyBase64 = base64.StdEncoding.EncodeToString(publicKeyBytes)
	resp.ValidationStatus = "trusted" // Assuming fedClient.GetServerTLSCertificate only returns trusted ones
	resp.FingerprintSHA256 = fmt.Sprintf("%x", sha256.Sum256(cert.Raw))
	
	logger.Info("Successfully retrieved and returned TLS certificate for remote server")
	return util.JSONResponse{
		Code: http.StatusOK,
		JSON: resp,
	}
}