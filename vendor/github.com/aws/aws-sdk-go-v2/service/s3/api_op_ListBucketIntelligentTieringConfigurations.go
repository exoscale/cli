// Code generated by smithy-go-codegen DO NOT EDIT.

package s3

import (
	"context"
	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	s3cust "github.com/aws/aws-sdk-go-v2/service/s3/internal/customizations"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

// Lists the S3 Intelligent-Tiering configuration from the specified bucket. The S3
// Intelligent-Tiering storage class is designed to optimize storage costs by
// automatically moving data to the most cost-effective storage access tier,
// without additional operational overhead. S3 Intelligent-Tiering delivers
// automatic cost savings by moving data between access tiers, when access patterns
// change. The S3 Intelligent-Tiering storage class is suitable for objects larger
// than 128 KB that you plan to store for at least 30 days. If the size of an
// object is less than 128 KB, it is not eligible for auto-tiering. Smaller objects
// can be stored, but they are always charged at the frequent access tier rates in
// the S3 Intelligent-Tiering storage class. If you delete an object before the end
// of the 30-day minimum storage duration period, you are charged for 30 days. For
// more information, see Storage class for automatically optimizing frequently and
// infrequently accessed objects
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/storage-class-intro.html#sc-dynamic-data-access).
// Operations related to ListBucketIntelligentTieringConfigurations include:
//
// *
// DeleteBucketIntelligentTieringConfiguration
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_DeleteBucketIntelligentTieringConfiguration.html)
//
// *
// PutBucketIntelligentTieringConfiguration
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutBucketIntelligentTieringConfiguration.html)
//
// *
// GetBucketIntelligentTieringConfiguration
// (https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetBucketIntelligentTieringConfiguration.html)
func (c *Client) ListBucketIntelligentTieringConfigurations(ctx context.Context, params *ListBucketIntelligentTieringConfigurationsInput, optFns ...func(*Options)) (*ListBucketIntelligentTieringConfigurationsOutput, error) {
	if params == nil {
		params = &ListBucketIntelligentTieringConfigurationsInput{}
	}

	result, metadata, err := c.invokeOperation(ctx, "ListBucketIntelligentTieringConfigurations", params, optFns, addOperationListBucketIntelligentTieringConfigurationsMiddlewares)
	if err != nil {
		return nil, err
	}

	out := result.(*ListBucketIntelligentTieringConfigurationsOutput)
	out.ResultMetadata = metadata
	return out, nil
}

type ListBucketIntelligentTieringConfigurationsInput struct {

	// The name of the Amazon S3 bucket whose configuration you want to modify or
	// retrieve.
	//
	// This member is required.
	Bucket *string

	// The ContinuationToken that represents a placeholder from where this request
	// should begin.
	ContinuationToken *string
}

type ListBucketIntelligentTieringConfigurationsOutput struct {

	// The ContinuationToken that represents a placeholder from where this request
	// should begin.
	ContinuationToken *string

	// The list of S3 Intelligent-Tiering configurations for a bucket.
	IntelligentTieringConfigurationList []types.IntelligentTieringConfiguration

	// Indicates whether the returned list of analytics configurations is complete. A
	// value of true indicates that the list is not complete and the
	// NextContinuationToken will be provided for a subsequent request.
	IsTruncated bool

	// The marker used to continue this inventory configuration listing. Use the
	// NextContinuationToken from this response to continue the listing in a subsequent
	// request. The continuation token is an opaque value that Amazon S3 understands.
	NextContinuationToken *string

	// Metadata pertaining to the operation's result.
	ResultMetadata middleware.Metadata
}

func addOperationListBucketIntelligentTieringConfigurationsMiddlewares(stack *middleware.Stack, options Options) (err error) {
	err = stack.Serialize.Add(&awsRestxml_serializeOpListBucketIntelligentTieringConfigurations{}, middleware.After)
	if err != nil {
		return err
	}
	err = stack.Deserialize.Add(&awsRestxml_deserializeOpListBucketIntelligentTieringConfigurations{}, middleware.After)
	if err != nil {
		return err
	}
	if err = addSetLoggerMiddleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddClientRequestIDMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddComputeContentLengthMiddleware(stack); err != nil {
		return err
	}
	if err = addResolveEndpointMiddleware(stack, options); err != nil {
		return err
	}
	if err = v4.AddComputePayloadSHA256Middleware(stack); err != nil {
		return err
	}
	if err = addRetryMiddlewares(stack, options); err != nil {
		return err
	}
	if err = addHTTPSignerV4Middleware(stack, options); err != nil {
		return err
	}
	if err = awsmiddleware.AddRawResponseToMetadata(stack); err != nil {
		return err
	}
	if err = awsmiddleware.AddRecordResponseTiming(stack); err != nil {
		return err
	}
	if err = addClientUserAgent(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddErrorCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = smithyhttp.AddCloseResponseBodyMiddleware(stack); err != nil {
		return err
	}
	if err = addOpListBucketIntelligentTieringConfigurationsValidationMiddleware(stack); err != nil {
		return err
	}
	if err = stack.Initialize.Add(newServiceMetadataMiddleware_opListBucketIntelligentTieringConfigurations(options.Region), middleware.Before); err != nil {
		return err
	}
	if err = addMetadataRetrieverMiddleware(stack); err != nil {
		return err
	}
	if err = addListBucketIntelligentTieringConfigurationsUpdateEndpoint(stack, options); err != nil {
		return err
	}
	if err = addResponseErrorMiddleware(stack); err != nil {
		return err
	}
	if err = v4.AddContentSHA256HeaderMiddleware(stack); err != nil {
		return err
	}
	if err = disableAcceptEncodingGzip(stack); err != nil {
		return err
	}
	if err = addRequestResponseLogging(stack, options); err != nil {
		return err
	}
	return nil
}

func newServiceMetadataMiddleware_opListBucketIntelligentTieringConfigurations(region string) *awsmiddleware.RegisterServiceMetadata {
	return &awsmiddleware.RegisterServiceMetadata{
		Region:        region,
		ServiceID:     ServiceID,
		SigningName:   "s3",
		OperationName: "ListBucketIntelligentTieringConfigurations",
	}
}

// getListBucketIntelligentTieringConfigurationsBucketMember returns a pointer to
// string denoting a provided bucket member valueand a boolean indicating if the
// input has a modeled bucket name,
func getListBucketIntelligentTieringConfigurationsBucketMember(input interface{}) (*string, bool) {
	in := input.(*ListBucketIntelligentTieringConfigurationsInput)
	if in.Bucket == nil {
		return nil, false
	}
	return in.Bucket, true
}
func addListBucketIntelligentTieringConfigurationsUpdateEndpoint(stack *middleware.Stack, options Options) error {
	return s3cust.UpdateEndpoint(stack, s3cust.UpdateEndpointOptions{
		Accessor: s3cust.UpdateEndpointParameterAccessor{
			GetBucketFromInput: getListBucketIntelligentTieringConfigurationsBucketMember,
		},
		UsePathStyle:            options.UsePathStyle,
		UseAccelerate:           options.UseAccelerate,
		SupportsAccelerate:      true,
		EndpointResolver:        options.EndpointResolver,
		EndpointResolverOptions: options.EndpointOptions,
		UseDualstack:            options.UseDualstack,
		UseARNRegion:            options.UseARNRegion,
	})
}
