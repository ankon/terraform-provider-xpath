package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
)

func nodeTypeToName(typ xmlquery.NodeType) string {
	switch typ {
	case xmlquery.DocumentNode:
		return "document"
	case xmlquery.DeclarationNode:
		return "declaration"
	case xmlquery.ElementNode:
		return "element"
	case xmlquery.TextNode:
		return "text"
	case xmlquery.CharDataNode:
		return "cdata"
	case xmlquery.CommentNode:
		return "comment"
	case xmlquery.AttributeNode:
		return "attribute"
	case xmlquery.NotationNode:
		return "notation"
	default:
		return fmt.Sprintf("unknown_%x", typ)
	}
}

// XPathDataSourceModel describes the data source data model.
type XPathDataSourceModel struct {
	Expression        types.String  `tfsdk:"expression"`
	Content           types.String  `tfsdk:"content"`
	ContentURL        types.String  `tfsdk:"content_url"`
	NamespaceBindings types.Map     `tfsdk:"namespace_bindings"`
	Result            types.Dynamic `tfsdk:"result"`
}

func xpathQueryHelper(ctx context.Context, d XPathDataSourceModel) (result []*xmlquery.Node, err error) {
	var document *xmlquery.Node
	if content := d.Content.ValueString(); content != "" {
		contentReader := strings.NewReader(content)
		d, err := xmlquery.Parse(contentReader)
		if err != nil {
			return nil, fmt.Errorf("cannot parse content: %w", err)
		}
		document = d
	} else if contentURL := d.ContentURL.ValueString(); contentURL != "" {
		d, err := xmlquery.LoadURL(contentURL)
		if err != nil {
			return nil, fmt.Errorf("cannot load/parse content: %v", err)
		}
		document = d
	}

	nsMap := map[string]string{}
	for prefix, value := range d.NamespaceBindings.Elements() {
		bindingValue, err := value.ToTerraformValue(ctx)
		if err != nil {
			return nil, fmt.Errorf("binding for prefix %s is not a terraform value", prefix)
		}

		var binding string
		err = bindingValue.As(&binding)
		if err != nil {
			return nil, fmt.Errorf("binding for prefix %s is not a string", prefix)
		}
		nsMap[prefix] = binding
	}

	expr, err := xpath.CompileWithNS(d.Expression.ValueString(), nsMap)
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	return xmlquery.QuerySelectorAll(document, expr), nil
}
