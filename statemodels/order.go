package statemodels

// Order is the state model of the admin/order state
type Order struct {
	OrderID uint
}

type VolunteerOrder struct {
	CurrentOrderID uint
	HasNext        bool
	HasPrevious    bool
}
