package helper

type AddressResponse struct {
	UserID   uint32 `json:"userId"`
	City     string `json:"city"`
	District string `json:"district"`
	State    string `json:"state"`
	Road     string `json:"road"`
}
