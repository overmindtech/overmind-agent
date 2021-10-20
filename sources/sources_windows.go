package sources

func init() {
	netstatSource := netstat.PortSource{}

	if netstatSource.Supported() {
		Sources = append(Sources, &netstatSource)
	}
}
