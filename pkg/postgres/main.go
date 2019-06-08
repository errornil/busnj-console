package postgres

type Client struct {

}

func NewClient() (*Client, error) {
	return &Client{

	}, nil
}

func (c *Client) GetTripByBlockIDAndHeadsign(blockID, headsign string) (Trip, error) {
	return Trip{
		
	}, nil
}
