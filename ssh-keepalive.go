package ssh

func (c *ClientConn) CheckServerAlive() error {
	_, err := c.sendGlobalRequest(globalRequestMsg{"keepalive@openssh.com", true})

	if err == nil || err.Error() == "request failed" {
		// Any response is a success.
		return nil
	}

	return err
}
