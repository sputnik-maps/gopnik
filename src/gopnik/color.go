package gopnik

type RGBAColor struct {
	R, G, B, A uint32
}

func (self RGBAColor) RGBA() (r, g, b, a uint32) {
	return self.R, self.G, self.B, self.A
}
