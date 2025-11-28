package ca

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

func CreateCA() {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization: []string{"VMware ESX Server Default Certificate"},
			Country:      []string{"US"},
			Province:     []string{"California"},
			Locality:     []string{"Palo Alto"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}

	// Public key
	certOut, err := os.Create("cert/ca.crt")
	if err != nil {
		logrus.Fatal(err)
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: ca_b})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not encode certificate file")
	}
	err = certOut.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not close certificate file")
	}
	logrus.WithFields(logrus.Fields{
		"cert": "ca.pem created",
	}).Info("cert")

	// Private key
	keyOut, err := os.OpenFile("cert/ca.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logrus.Fatal(err)
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not encode private key file")
	}
	err = keyOut.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not close private key file")
	}
	logrus.WithFields(logrus.Fields{
		"cert": "ca.key created",
	}).Info("cert")
}

func CreateCert(path string, name string, cn string) {

	// Load CA
	catls, err := tls.LoadX509KeyPair("cert/ca.crt", "cert/ca.key")
	if err != nil {
		logrus.Fatal(err)
	}
	ca, err := x509.ParseCertificate(catls.Certificate[0])
	if err != nil {
		logrus.Fatal(err)
	}

	// Prepare certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Organization: []string{"VMware ESX Server Default Certificate"},
			Country:      []string{"US"},
			Province:     []string{"California"},
			Locality:     []string{"Palo Alto"},
			CommonName:   cn,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	// add SAN
	cert.DNSNames = []string{cn}
	cert.EmailAddresses = []string{"ssl-certificates@vmware.com"}

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey

	// Sign the certificate
	cert_b, err := x509.CreateCertificate(rand.Reader, cert, ca, pub, catls.PrivateKey)
	if err != nil {
		logrus.Fatal(err)
	}

	// Public key
	certOut, err := os.Create(path + "/" + name + ".crt")
	if err != nil {
		logrus.Fatal(err)
	}
	err = pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: cert_b})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not encode certificate file")
	}
	err = certOut.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not close certificate file")
	}
	logrus.WithFields(logrus.Fields{
		"cert": path + "/" + name + ".crt created",
	}).Info("cert")

	// Private key
	keyOut, err := os.OpenFile(path+"/"+name+".key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		logrus.Fatal(err)
	}
	err = pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not encode private key file")
	}
	err = keyOut.Close()
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err": err,
		}).Warn("could not close private key file")
	}
	logrus.WithFields(logrus.Fields{
		"cert": path + "/" + name + ".key created",
	}).Info("cert")

}
