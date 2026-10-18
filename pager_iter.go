//go:build go1.23

package voiceml

import "iter"

// All adapts the Pager to a range-over-func iterator (Go 1.23+). It yields
// each remaining item with a nil error; if fetching a page fails it yields
// the zero value and the error once, then stops. Breaking out of the loop
// stops further requests.
//
//	for call, err := range c.Calls.Iterate(ctx, params).All() {
//		if err != nil {
//			return err
//		}
//		fmt.Println(call.Sid)
//	}
func (p *Pager[T]) All() iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for p.Next() {
			if !yield(p.Current(), nil) {
				return
			}
		}
		if err := p.Err(); err != nil {
			var zero T
			yield(zero, err)
		}
	}
}
