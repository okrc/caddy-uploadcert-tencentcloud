package uploadcert_tencentcloud

import (
	"context"
	"fmt"
	"slices"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/certmagic"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func init() {
	caddy.RegisterModule(TencentCloudCertHandler{})
}

type TencentCloudCertHandler struct {
	SecretId  string   `json:"secret_id"`
	SecretKey string   `json:"secret_key"`
	AllowList []string `json:"allow_list,omitempty"`
	BlockList []string `json:"block_list,omitempty"`

	storage certmagic.Storage
	logger  *zap.Logger
}

func (TencentCloudCertHandler) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "events.handlers.upload_cert_tencentcloud",
		New: func() caddy.Module { return new(TencentCloudCertHandler) },
	}
}

func (h *TencentCloudCertHandler) Provision(ctx caddy.Context) error {
	h.logger = ctx.Logger(h)
	h.storage = ctx.Storage()
	return nil
}

func (h *TencentCloudCertHandler) Handle(ctx context.Context, e caddy.Event) error {
	if e.Name() != "cert_obtained" {
		h.logger.Warn("upload_cert_tencentcloud should only be handled on `cert_obtained`, ignoring", zap.String("event", e.Name()))
		return nil
	}
	certID, ok := e.Data["identifier"].(string)
	if !ok {
		return errors.New("missing certificate identifier")
	}
	if slices.Contains(h.BlockList, certID) || len(h.AllowList) > 0 && !slices.Contains(h.AllowList, certID) {
		h.logger.Warn(fmt.Sprintf("upload_cert_tencentcloud ignored certificate %s not matching the current rule", certID), zap.String("event", e.Name()))
		return nil
	}
	certificatePath, ok := e.Data["certificate_path"].(string)
	if !ok {
		return errors.New("missing certificate path")
	}
	privateKeyPath, ok := e.Data["private_key_path"].(string)
	if !ok {
		return errors.New("missing private key path")
	}

	loadCert := func(path string) (string, error) {
		bytes, err := h.storage.Load(ctx, path)
		if err != nil {
			return "", errors.Wrapf(err, "failed to load file: %s", path)
		}
		return string(bytes), nil
	}

	cert, err := loadCert(certificatePath)
	if err != nil {
		return err
	}
	key, err := loadCert(privateKeyPath)
	if err != nil {
		return err
	}

	go func() {
		certificateId, _ := h.DescribeCertificates(ctx, certID)
		if certificateId == "" {
			if err := h.UploadCertificate(ctx, cert, key); err != nil {
				h.logger.Error("failed to upload certificate", zap.Error(err))
			}
		} else {
			if err := h.UpdateCertificateInstance(ctx, cert, key, certificateId); err != nil {
				h.logger.Error("failed to update certificate instance", zap.Error(err))
			}
		}
	}()
	h.logger.Info(fmt.Sprintf("upload_cert_tencentcloud successfully uploaded domain %s to Tencent Cloud", certID), zap.String("event", e.Name()))
	return nil
}

func (h *TencentCloudCertHandler) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	for d.Next() {
		if d.NextArg() {
			return d.ArgErr()
		}
		for nesting := d.Nesting(); d.NextBlock(nesting); {
			switch d.Val() {
			case "secret_id":
				if !d.NextArg() {
					return d.ArgErr()
				}
				h.SecretId = d.Val()
			case "secret_key":
				if !d.NextArg() {
					return d.ArgErr()
				}
				h.SecretKey = d.Val()
			case "allow_list":
				for d.NextArg() {
					h.AllowList = append(h.AllowList, d.Val())
				}
			case "block_list":
				for d.NextArg() {
					h.BlockList = append(h.BlockList, d.Val())
				}
			default:
				return d.Errf("unrecognized subdirective '%s'", d.Val())
			}
			if d.NextArg() {
				return d.ArgErr()
			}
		}
	}
	if h.SecretId == "" || h.SecretKey == "" {
		return d.Err("SecretId or SecretKey is empty")
	}
	return nil
}

// Interface guards
var (
	_ caddyfile.Unmarshaler = (*TencentCloudCertHandler)(nil)
	_ caddy.Provisioner     = (*TencentCloudCertHandler)(nil)
)
