package uidev

type ButtonVariant string

const (
	ButtonVariantOutline ButtonVariant = "outline"
	ButtonVariantGhost   ButtonVariant = "ghost"
)

// ButtonSize defines the size variant of the button
type ButtonSize string

const (
	ButtonSizeSmall ButtonSize = "small"
	ButtonSizeLarge ButtonSize = "large"
)
