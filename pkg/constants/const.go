package constants

const SocketPathInContainer = "/cid/socket/cid.socket"
const TempPathInContainer = "/cid/temp"

// CaBundles are taken from https://go.dev/src/crypto/x509/root_linux.go
var CaBundles = [][]string{
	{"/etc/ssl/certs/ca-certificates.crt"},                                  // Debian/Ubuntu/Gentoo etc.
	{"/etc/pki/tls/certs/ca-bundle.crt", "/etc/pki/tls/certs/ca-extra.crt"}, // Fedora/RHEL 6
	{"/etc/ssl/ca-bundle.pem"},                                              // OpenSUSE
	{"/etc/pki/tls/cacert.pem"},                                             // OpenELEC
	{"/etc/pki/ca-trust/extracted/pem/tls-ca-bundle.pem"},                   // CentOS/RHEL 7
	{"/etc/ssl/cert.pem"},                                                   // Alpine Linux
}
