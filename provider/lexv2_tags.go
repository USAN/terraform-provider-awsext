package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func readLexV2Tags(ctx context.Context, client *lexmodelsv2.Client, resourceArn string) (types.Map, error) {
	out, err := client.ListTagsForResource(ctx, &lexmodelsv2.ListTagsForResourceInput{
		ResourceARN: aws.String(resourceArn),
	})
	if err != nil {
		return types.MapNull(types.StringType), err
	}

	result, diags := types.MapValueFrom(ctx, types.StringType, out.Tags)
	if diags.HasError() {
		return types.MapNull(types.StringType), fmt.Errorf("error converting lex v2 tags to map")
	}
	return result, nil
}

func applyLexV2Tags(ctx context.Context, client *lexmodelsv2.Client, resourceArn string, tags types.Map) error {
	if tags.IsNull() || tags.IsUnknown() {
		return nil
	}
	var tagMap map[string]string
	if diags := tags.ElementsAs(ctx, &tagMap, false); diags.HasError() {
		return fmt.Errorf("error reading lex v2 tags")
	}
	if len(tagMap) == 0 {
		return nil
	}
	_, err := client.TagResource(ctx, &lexmodelsv2.TagResourceInput{
		ResourceARN: aws.String(resourceArn),
		Tags:        tagMap,
	})
	return err
}

func updateLexV2Tags(ctx context.Context, client *lexmodelsv2.Client, resourceArn string, oldTags, newTags types.Map) error {
	var oldTagMap, newTagMap map[string]string

	if !oldTags.IsNull() && !oldTags.IsUnknown() {
		oldTags.ElementsAs(ctx, &oldTagMap, false)
	}
	if !newTags.IsNull() && !newTags.IsUnknown() {
		newTags.ElementsAs(ctx, &newTagMap, false)
	}

	var keysToDelete []string
	for k := range oldTagMap {
		if _, exists := newTagMap[k]; !exists {
			keysToDelete = append(keysToDelete, k)
		}
	}

	if len(keysToDelete) > 0 {
		if _, err := client.UntagResource(ctx, &lexmodelsv2.UntagResourceInput{
			ResourceARN: aws.String(resourceArn),
			TagKeys:     keysToDelete,
		}); err != nil {
			return err
		}
	}

	if len(newTagMap) > 0 {
		if _, err := client.TagResource(ctx, &lexmodelsv2.TagResourceInput{
			ResourceARN: aws.String(resourceArn),
			Tags:        newTagMap,
		}); err != nil {
			return err
		}
	}

	return nil
}
