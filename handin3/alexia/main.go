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
	key := []byte("1098765432100001")
	n,d:=KeyGen(16)

	//key in RSA (n,e)
	pk := generatePublicKey(n,e)
	fmt.Println("RSA Public Key (n,e): ",pk.N_pk, pk.E_pk)
	sk := generateSecretKey(n,d)
	fmt.Println("RSA Secret Key (n,d): ",sk.N_sk,sk.D_sk)
	
	var buffer bytes.Buffer
	var writer = bufio.NewWriter(&buffer)	
	var encoder = gob.NewEncoder(writer)
	err := encoder.Encode(&sk)
	if (err != nil){
		panic(err.Error())
	} else{
		writer.Flush()
	}
	//fmt.Println("Encoded RSA secret key: ",buffer)

	//cipherkey := Encrypt(big.NewInt(15111),n)
	//fmt.Println("RSA ciphertext key: ",cipherkey)
	fmt.Println("Encrypting RSA secret key to file...")
	EncryptToFile(buffer.Bytes(),key)
	encoded_key := DecryptFromFile(key)
	fmt.Println("Decrypted encoded RSA key: ",encoded_key)

	buffer.Reset()
	buffer.Write(encoded_key)
	var reader = bufio.NewReader(&buffer)
	var secretKey = &SecretKey{}
	var dec = gob.NewDecoder(reader)
	err = dec.Decode(secretKey)
	if (err != nil){
		fmt.Println("Error - ", err.Error())
	}
	fmt.Println("Decrypted RSA secret key: ", secretKey.N_sk,secretKey.D_sk)
}