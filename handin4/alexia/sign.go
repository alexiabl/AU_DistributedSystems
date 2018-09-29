package main

import (
	"crypto/sha256"
)

func hash(message []byte) ([]byte) {
	hash := sha256.New()
	hash.Write(message)
	return hash.Sum(nil)
}

func sign(){

}
/*
func main(){
	hash := hash([]byte("Hi my name is Alexia"))
	fmt.Printf("%x",hash)
}*/