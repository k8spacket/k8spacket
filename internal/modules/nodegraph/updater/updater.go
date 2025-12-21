package updater

type Updater interface {
	Update(src string, srcName string, srcNamespace string, dst string, dstName string, dstNamespace string, persistent bool, bytesSent float64, bytesReceived float64, duration float64, closed bool)
}
