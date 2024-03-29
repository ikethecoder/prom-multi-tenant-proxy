package injector

import (
	"fmt"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql/parser"
)

//SetRecursive Injects the label in the Prometheus Query Node
func SetRecursive(node parser.Node, matchersToEnforce []*labels.Matcher) (err error) {
	switch n := node.(type) {
	case *parser.EvalStmt:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case parser.Expressions:
		for _, e := range n {
			if err := SetRecursive(e, matchersToEnforce); err != nil {
				return err
			}
		}
	case *parser.AggregateExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.BinaryExpr:
		if err := SetRecursive(n.LHS, matchersToEnforce); err != nil {
			return err
		}
		if err := SetRecursive(n.RHS, matchersToEnforce); err != nil {
			return err
		}

	case *parser.Call:
		if err := SetRecursive(n.Args, matchersToEnforce); err != nil {
			return err
		}

	case *parser.ParenExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.UnaryExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.SubqueryExpr:
		if err := SetRecursive(n.Expr, matchersToEnforce); err != nil {
			return err
		}

	case *parser.NumberLiteral, *parser.StringLiteral:
	// nothing to do

	case *parser.MatrixSelector:
		// inject labelselector
		n.VectorSelector.(*parser.VectorSelector).LabelMatchers = enforceLabelMatchers(n.VectorSelector.(*parser.VectorSelector).LabelMatchers, matchersToEnforce)

	case *parser.VectorSelector:
		// inject labelselector
		n.LabelMatchers = enforceLabelMatchers(n.LabelMatchers, matchersToEnforce)

	default:
		panic(fmt.Errorf("parser.Walk: unhandled node type %T", node))
	}

	return err
}

func enforceLabelMatchers(matchers []*labels.Matcher, matchersToEnforce []*labels.Matcher) []*labels.Matcher {
	res := []*labels.Matcher{}
	for _, m := range matchersToEnforce {
		res = enforceLabelMatcher(matchers, m)
	}

	return res
}

func enforceLabelMatcher(matchers []*labels.Matcher, enforcedMatcher *labels.Matcher) []*labels.Matcher {
	res := []*labels.Matcher{}
	for _, m := range matchers {
		if m.Name == enforcedMatcher.Name {
			// do not exclude - we are adding to the query
			// continue
		}
		res = append(res, m)
	}

	return append(res, enforcedMatcher)
}
