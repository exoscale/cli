package entities

import (
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type ObjectListing struct {
	List           []types.Object
	CommonPrefixes []string
}
