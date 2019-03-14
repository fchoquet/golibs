package pager

// Pager represents a pager used to return results by page
// it is 0-based for pages and items
// Limit = 0 means no limit so a zeroed Pager means no pagination
type Pager struct {
	Page  int
	Limit int
}

// New returns a new pager
func New(page int, limit int) Pager {
	return Pager{
		Page:  page,
		Limit: limit,
	}
}

// NoPager builds a valid pager corresponding to "no pager"
// ie: a pager that returns all the results in one page
func NoPager() Pager {
	return Pager{}
}

// Enabled returns true if pager is enabled
func (p Pager) Enabled() bool {
	return p.Limit > 0
}

// First return index of first element in the page
func (p Pager) First() int {
	return p.Limit * p.Page
}

// Last return index of last element in the page
func (p Pager) Last() int {
	if !p.Enabled() {
		return 0
	}

	return p.Limit*(p.Page+1) - 1
}

// IsVisible returns true if given index is visible in current page
func (p Pager) IsVisible(index int) bool {
	if !p.Enabled() {
		return true
	}

	return index >= p.First() && index <= p.Last()
}

// PageOf returns the number of the page where a given index is located
func (p Pager) PageOf(index int) int {
	if !p.Enabled() {
		return 0
	}

	return index / p.Limit
}
