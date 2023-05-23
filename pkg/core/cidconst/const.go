package cidconst

const SocketPathInContainer = "/cid/socket/cid.socket"
const TempPathInContainer = "/cid/temp"

// see https://go.dev/src/crypto/x509/root_linux.go for possible paths
var CaBundles = [][]string{
	{"/etc/ssl/certs/ca-certificates.crt"},                                  // Debian/Ubuntu/Gentoo etc.
	{"/etc/pki/tls/certs/ca-bundle.crt", "/etc/pki/tls/certs/ca-extra.crt"}, // RHEL
	{"/etc/ssl/ca-bundle.pem"},                                              // OpenSUSE
	{"/etc/pki/tls/cacert.pem"},                                             // OpenELEC
	{"/etc/ssl/cert.pem"},                                                   // Alpine Linux
}
