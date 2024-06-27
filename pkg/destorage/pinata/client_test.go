package pinata_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stafiprotocol/eth-lsd-relay/pkg/destorage/pinata"
	"github.com/stretchr/testify/assert"
)

func TestUploadAndDownload(t *testing.T) {
	pinataJwt := os.Getenv("PINATA_JWT")
	pinataEndpoint := os.Getenv("PINATA_ENDPOINT")
	client, err := pinata.NewClient(pinataEndpoint, pinataJwt)
	assert.Nil(t, err)

	fileName := "hello-world.txt"
	fileContent := []byte("你好世界！Hello World! - " + time.Now().String())
	cid, err := client.UploadFile(fileContent, fileName)
	assert.Nil(t, err)
	assert.NotEmpty(t, cid)
	fmt.Println("file name:", fileName)
	fmt.Println("cid:", cid)

	downloadContent, err := client.DownloadFile(cid, fileName)
	assert.Nil(t, err)
	assert.Equal(t, string(fileContent), string(downloadContent))
	fmt.Println("content:", string(downloadContent))
}

func TestUnpinFilesCreatedBefore(t *testing.T) {
	pinataJwt := os.Getenv("PINATA_JWT")
	pinataEndpoint := os.Getenv("PINATA_ENDPOINT")
	client, err := pinata.NewClient(pinataEndpoint, pinataJwt)
	assert.Nil(t, err)

	count, err := client.UnpinFilesCreatedBefore(time.Now().AddDate(0, -180, 0))
	assert.Nil(t, err)
	fmt.Println("deleted count:", count)
}
