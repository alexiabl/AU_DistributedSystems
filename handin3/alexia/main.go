package main

import (
	"fmt"
	"encoding/gob"
	"bytes"
	"bufio"
)

func main(){
	fmt.Println("RSA TEST:")
	fmt.Println("=============")
	testRSA()

	fmt.Println("\nAES TEST:")
	fmt.Println("=============")
	testAES()

	fmt.Println("\nTESTING RSA WITH AES:")
	fmt.Println("=============")
	key := []byte("1098765432100000")
	n,p,q:=KeyGen(14)

	//key in RSA (n,e)
	pk := generatePublicKey(n,e)
	fmt.Println("Public Key (n,e): ",pk.N_pk, pk.E_pk)
	d := calculateD(p,q)
	sk := generateSecretKey(n,d)
	fmt.Println("Secret Key (n,d): ",sk.N_sk,sk.D_sk)
	
	var buffer bytes.Buffer
	var writer = bufio.NewWriter(&buffer)	
	var encoder = gob.NewEncoder(writer)
	err := encoder.Encode(&pk)
	if (err != nil){
		panic(err.Error())
	} else{
		writer.Flush()
	}
	fmt.Println("Encoded public key: ",buffer)
	EncryptToFile(buffer.Bytes(),key)
	encoded_key := DecryptFromFile(key)
	fmt.Println("Decrypted encoded key: ",encoded_key)

	buffer.Reset()
	buffer.Write(encoded_key)
	var reader = bufio.NewReader(&buffer)
	var publicKey = &PublicKey{}
	var dec = gob.NewDecoder(reader)
	err = dec.Decode(publicKey)
	if (err != nil){
		fmt.Println("Error - ", err.Error())
	}
	fmt.Println("Decrypted key: ", publicKey.N_pk,publicKey.E_pk)
}