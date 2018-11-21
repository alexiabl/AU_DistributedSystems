package main

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"strconv"
)

func (c *Client) CalculateDrawValue(seed int, slot int, draw *big.Int, publicKey PublicKey) *big.Int {
	_, drawMsgStr := GenerateDrawMessage(seed, slot)

	shaMsg := []byte(drawMsgStr + publicKey.toString() + draw.String())
	sha := sha256.New()
	sha.Write(shaMsg)
	hash := sha.Sum(nil)

	val := new(big.Int).Mul(big.NewInt(int64(PREMIUM_ACCOUNT)), new(big.Int).SetBytes(hash))
	return val
}

func (c *Client) IsValidDraw(seed int, slot int, draw *big.Int, senderPk PublicKey) bool {
	// Make sure it's a king who signed the message
	isKing := false
	for _, key := range c.genesisBlock.KingKeys {
		if senderPk.toString() == key {
			isKing = true
			break
		}
	}
	if !isKing {
		fmt.Println("Invalid draw: Isn't king")
		return false
	}

	// Make sure that the value is above the hardness
	val := c.CalculateDrawValue(seed, slot, draw, senderPk)
	if val.Cmp(HARDNESS) < 0 {
		fmt.Println("Invalid draw: the value is too low")
		return false
	}

	drawMsg, _ := GenerateDrawMessage(seed, slot)
	valid := Verify(drawMsg, draw, senderPk)

	if !valid {
		fmt.Println("Invalid draw: could verify the draw with the signature")
	}

	return valid
}

func GenerateDrawMessage(seed int, slot int) ([]byte, string) {
	drawMsgStr := strconv.Itoa(seed + slot)
	drawMsg := []byte(drawMsgStr)
	return drawMsg, drawMsgStr
}

func GenerateDraw(seed int, slot int, sk SecretKey) *big.Int {
	drawMsg, _ := GenerateDrawMessage(seed, slot)
	draw := Sign(drawMsg, sk)
	return draw
}
