package services

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

var errForbiddenTarget = errors.New("endereço bloqueado: aponta para rede privada ou reservada")

// validateFetchURL garante que a URL possa ser buscada com segurança pelo
// servidor, bloqueando endereços de rede privada/loopback/link-local e
// esquemas não-HTTP (mitigação de SSRF). Hosts são resolvidos por DNS e todos
// os IPs retornados são verificados.
func validateFetchURL(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("URL inválida: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("URL inválida: apenas http/https são permitidos")
	}

	host := parsed.Hostname()
	if host == "" {
		return fmt.Errorf("URL inválida: host vazio")
	}

	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: %s", errForbiddenTarget, host)
		}
		return nil
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("falha ao resolver host %s: %w", host, err)
	}
	if len(ips) == 0 {
		return fmt.Errorf("falha ao resolver host %s: nenhum endereço encontrado", host)
	}
	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: %s (%s)", errForbiddenTarget, host, ip.String())
		}
	}
	return nil
}

// isBlockedIP cobre as faixas reservadas/privadas/proibidas em IPv4 e IPv6,
// incluindo aquelas não cobertas pelos helpers padrão de net.IP.
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}

	if v4 := ip.To4(); v4 != nil {
		// 100.64.0.0/10 (CGNAT) e 0.0.0.0/8 não são cobertos pelos helpers padrão.
		if (v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127) || v4[0] == 0 {
			return true
		}
		return v4.IsLoopback() || v4.IsPrivate() || v4.IsLinkLocalUnicast() ||
			v4.IsLinkLocalMulticast() || v4.IsUnspecified()
	}

	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() || ip.IsUnspecified()
}
