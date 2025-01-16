package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
)

const m = 8 // 8 bits, range 0-255

var two = big.NewInt(2)
var hashMod = new(big.Int).Exp(two, big.NewInt(m), nil)

func hashString(data string) *big.Int {
	hasher := sha1.New()
	hasher.Write([]byte(data))
	return new(big.Int).SetBytes(hasher.Sum(nil))
}

// IdentifierGen Generate an identifier for a given IP address
func IdentifierGen(IPAdd string) *big.Int {
	var identifier *big.Int
	identifier = hashString(IPAdd)
	identifier.Mod(identifier, hashMod)
	return identifier
}

// FolderPathGen generates the folder path directory for the given node's ID
func FolderPathGen(NodeId *big.Int) string {
	FolderName := "../Node_files/" + "N" + NodeId.String()
	return FolderName
}

func readStringIn(input *bufio.Reader) string {
	file, _ := input.ReadString('\n')
	file = strings.TrimSpace(file)
	return file
}

func module(value *big.Int) *big.Int {
	return new(big.Int).Mod(value, hashMod)
}
func between(start, elt, end *big.Int, inclusive bool) bool {
	if end.Cmp(start) > 0 {
		return (start.Cmp(elt) < 0 && elt.Cmp(end) < 0) || (inclusive && elt.Cmp(end) == 0)
	} else {
		return start.Cmp(elt) < 0 || elt.Cmp(end) < 0 || (inclusive && elt.Cmp(end) == 0)
	}
}

func jump(address string, fingerentry int) *big.Int {
	n := IdentifierGen(address)
	fingerentryminus1 := big.NewInt(int64(fingerentry) - 1)
	jump := new(big.Int).Exp(two, fingerentryminus1, nil)
	sum := new(big.Int).Add(n, jump)

	return new(big.Int).Mod(sum, hashMod)
}

func getLocalAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String()
}
func GetPublicIP() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(body)
}

// RemoveFile removes a file from a node's folder, by giving the nodeId and the fileName.
func RemoveFile(nodeId *big.Int, file string) error {
	OldFileP := FolderPathGen(nodeId) + "/" + file
	err := os.Remove(OldFileP)
	if err != nil {
		log.Fatal("Error Removing the file:", err)
		return err
	}
	return nil
}

// GenAsymKeys generates asymmetric keys based on the RSA Asymmetric Cryptography.
func GenAsymKeys() (*rsa.PublicKey, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Printf("Error generating RSA private key: %s", err)
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return publicKey, privateKey, nil
}

func (cr *ChordRing) genRSAKey(bits int) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		log.Println("[genRSAKey] Failed to generate private key for node ", "N"+cr.Identifier.String())
	}
	cr.PrivateKey = privateKey
	cr.PublicKey = &privateKey.PublicKey

	//store private key in the node folder
	privateKeyDER := x509.MarshalPKCS1PrivateKey(privateKey)
	block := pem.Block{Type: "N" + cr.Identifier.String() + "-private Key",
		Headers: nil,
		Bytes:   privateKeyDER}
	//nodeFolder := "../files/" + "N" + cr.Identifier.String()
	privateKeyFile, err := os.Create("private.pem")
	if err != nil {
		log.Println("[genRSAKey] Failed to create private key file for node ", "N"+cr.Identifier.String())
	}
	defer privateKeyFile.Close()
	err = pem.Encode(privateKeyFile, &block)
	if err != nil {
		log.Println("[genRSAKey] Failed to write private key into file")
	}

	//store public kay in the node folder
	publicKeyDER, err := x509.MarshalPKIXPublicKey(cr.PublicKey)
	if err != nil {
		log.Println("[genRSAKey] Failed to get DER format of public key for node ", cr.Identifier)
	}
	block = pem.Block{
		Type:    "N" + cr.Identifier.String() + "-public Key",
		Headers: nil,
		Bytes:   publicKeyDER,
	}
	publicKeyFile, err := os.Create("public.pem")
	if err != nil {
		log.Println("[genRSAKey] Failed to create public key file for node ", cr.Identifier)
	}
	defer publicKeyFile.Close()
	err = pem.Encode(publicKeyFile, &block)
	if err != nil {
		log.Println("[genRSAKey] Failed to write public key into file")
	}
}
