package proxy

// for testing
func HelloProxy(sayhi string) (string, error) {
	returnMsg := "You sent: [" + sayhi + "]. We're testing proxy mode of OBD."
	return returnMsg, nil
}