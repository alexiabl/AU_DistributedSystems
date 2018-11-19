package main

import (
	"math/big"
	"strconv"
)

func IsValidDraw(seed int, slot int, signature *big.Int, pk PublicKey) bool {
	drawMsg := GenerateDrawMessage(seed, slot)
	return Verify(drawMsg, signature, pk)
}

func GenerateDrawMessage(seed int, slot int) []byte {
	drawMsgStr := strconv.Itoa(seed + slot)
	drawMsg := []byte(drawMsgStr)
	return drawMsg
}

func CalculateDraw(seed int, slot int, sk SecretKey) *big.Int {
	drawMsg := GenerateDrawMessage(seed, slot)
	draw := Sign(drawMsg, sk)
	return draw
}
