package processor

import (
	"image"
	"image/color"
)

type Blurstack struct {
	r, g, b, a uint32
	next *Blurstack
}

func (bs *Blurstack) NewBlurStack() *Blurstack {
	return &Blurstack{bs.r, bs.g, bs.b, bs.a, bs.next}
}

func Process(src image.Image, width, height, radius uint32, done chan<- struct{}) image.Image {
	var stackEnd, stackIn, stackOut *Blurstack
	var (
		div, widthMinus1, heightMinus1, radiusPlus1, sumFactor uint32
		x, y, i, p, yp, yi, yw,
		rSum, gSum, bSum, aSum,
		rOutSum, gOutSum, bOutSum, aOutSum,
		rInSum, gInSum, bInSum, aInSum,
		pr, pg, pb, pa uint32
	)

	img := toNRGBA(src)

	// log.Println("Image converted")

	div 		= radius + radius + 1
	widthMinus1 	= width - 1
	heightMinus1 	= height - 1
	radiusPlus1 	= radius + 1
	sumFactor 	= radiusPlus1 * (radiusPlus1 + 1) / 2

	bs := Blurstack{}
	stackStart := bs.NewBlurStack()
	stack := stackStart

	for i = 1; i < div; i++ {
		stack.next = bs.NewBlurStack()
		stack = stack.next
		if i == radiusPlus1 {
			stackEnd = stack
		}
	}
	stack.next = stackStart

	// log.Println("Image blur stacked")

	mulSum := mulTable[radius]
	shgSum := shgTable[radius]

	for y = 0; y < height; y++ {
		rInSum, gInSum, bInSum, aInSum, rSum, gSum, bSum, aSum = 0, 0, 0, 0, 0, 0, 0, 0

		pr = uint32(img.Pix[yi])
		pg = uint32(img.Pix[yi+1])
		pb = uint32(img.Pix[yi+2])
		pa = uint32(img.Pix[yi+3])

		rOutSum = radiusPlus1 * pr
		gOutSum = radiusPlus1 * pg
		bOutSum = radiusPlus1 * pb
		aOutSum = radiusPlus1 * pa

		rSum += sumFactor * pr
		gSum += sumFactor * pg
		bSum += sumFactor * pb
		aSum += sumFactor * pa

		stack = stackStart

		for i = 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa
			stack = stack.next
		}

		for i = 1; i < radiusPlus1; i++ {
			var diff uint32
			if widthMinus1 < i {
				diff = widthMinus1
			} else {
				diff = i
			}
			p = yi + (diff << 2)
			pr = uint32(img.Pix[p])
			pg = uint32(img.Pix[p+1])
			pb = uint32(img.Pix[p+2])
			pa = uint32(img.Pix[p+3])

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			rSum += stack.r * (radiusPlus1 - i)
			gSum += stack.g * (radiusPlus1 - i)
			bSum += stack.b * (radiusPlus1 - i)
			aSum += stack.a * (radiusPlus1 - i)

			rInSum += pr
			gInSum += pg
			bInSum += pb
			aInSum += pa

			stack = stack.next
		}
		stackIn = stackStart
		stackOut = stackEnd

		for x = 0; x < width; x++ {
			pa = (aSum * mulSum) >> shgSum
			img.Pix[yi+3] = uint8(pa)

			if pa != 0 {
				pa = 255 / pa
				img.Pix[yi]   = uint8((rSum * mulSum) >> shgSum)
				img.Pix[yi+1] = uint8((gSum * mulSum) >> shgSum)
				img.Pix[yi+2] = uint8((bSum * mulSum) >> shgSum)
			} else {
				img.Pix[yi]   = 0
				img.Pix[yi+1] = 0
				img.Pix[yi+2] = 0
			}

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum
			aSum -= aOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b
			aOutSum -= stackIn.a

			p = x + radius + 1

			if p > widthMinus1 {
				p = widthMinus1
			}
			p = (yw + p) << 2

			stackIn.r = uint32(img.Pix[p])
			stackIn.g = uint32(img.Pix[p+1])
			stackIn.b = uint32(img.Pix[p+2])
			stackIn.a = uint32(img.Pix[p+3])

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b
			aInSum += stackIn.a

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum
			aSum += aInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb
			aOutSum += pa

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb
			aInSum -= pa

			stackOut = stackOut.next

			yi += 4
		}
		yw += width
	}

	// log.Println("Blur height calculated")

	for x = 0; x < width; x++ {
		rInSum, gInSum, bInSum, aInSum, rSum, gSum, bSum, aSum = 0, 0, 0, 0, 0, 0, 0, 0

		yi = x << 2
		pr = uint32(img.Pix[yi])
		pg = uint32(img.Pix[yi+1])
		pb = uint32(img.Pix[yi+2])
		pa = uint32(img.Pix[yi+3])

		rOutSum = radiusPlus1 * pr
		gOutSum = radiusPlus1 * pg
		bOutSum = radiusPlus1 * pb
		aOutSum = radiusPlus1 * pa

		rSum += sumFactor * pr
		gSum += sumFactor * pg
		bSum += sumFactor * pb
		aSum += sumFactor * pa

		stack = stackStart

		for i = 0; i < radiusPlus1; i++ {
			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa
			stack = stack.next
		}

		yp = width

		for i = 1; i <= radius; i++ {
			yi = (yp + x) << 2
			pr = uint32(img.Pix[yi])
			pg = uint32(img.Pix[yi+1])
			pb = uint32(img.Pix[yi+2])
			pa = uint32(img.Pix[yi+3])

			stack.r = pr
			stack.g = pg
			stack.b = pb
			stack.a = pa

			rSum += stack.r * (radiusPlus1 - i)
			gSum += stack.g * (radiusPlus1 - i)
			bSum += stack.b * (radiusPlus1 - i)
			aSum += stack.a * (radiusPlus1 - i)

			rInSum += pr
			gInSum += pg
			bInSum += pb
			aInSum += pa

			stack = stack.next

			if i < heightMinus1 {
				yp += width
			}
		}

		yi = x
		stackIn = stackStart
		stackOut = stackEnd

		for y = 0; y < height; y++ {
			p = yi << 2
			pa = (aSum * mulSum) >> shgSum
			img.Pix[p+3] = uint8(pa)

			if pa > 0 {
				pa = 255 / pa
				img.Pix[p]   = uint8((rSum * mulSum) >> shgSum)
				img.Pix[p+1] = uint8((gSum * mulSum) >> shgSum)
				img.Pix[p+2] = uint8((bSum * mulSum) >> shgSum)
			} else {
				img.Pix[p]   = 0
				img.Pix[p+1] = 0
				img.Pix[p+2] = 0
			}

			rSum -= rOutSum
			gSum -= gOutSum
			bSum -= bOutSum
			aSum -= aOutSum

			rOutSum -= stackIn.r
			gOutSum -= stackIn.g
			bOutSum -= stackIn.b
			aOutSum -= stackIn.a

			p = y + radiusPlus1

			if p > heightMinus1 {
				p = heightMinus1
			}
			p = (x + (p * width)) << 2

			stackIn.r = uint32(img.Pix[p])
			stackIn.g = uint32(img.Pix[p+1])
			stackIn.b = uint32(img.Pix[p+2])
			stackIn.a = uint32(img.Pix[p+3])

			rInSum += stackIn.r
			gInSum += stackIn.g
			bInSum += stackIn.b
			aInSum += stackIn.a

			rSum += rInSum
			gSum += gInSum
			bSum += bInSum
			aSum += aInSum

			stackIn = stackIn.next

			pr = stackOut.r
			pg = stackOut.g
			pb = stackOut.b
			pa = stackOut.a

			rOutSum += pr
			gOutSum += pg
			bOutSum += pb
			aOutSum += pa

			rInSum -= pr
			gInSum -= pg
			bInSum -= pb
			aInSum -= pa

			stackOut = stackOut.next

			yi += width
		}
	}

	// log.Println("Blur width calculated")
	done <- struct{}{}
	return img
}

func toNRGBA(img image.Image) *image.NRGBA {
	srcBounds := img.Bounds()
	if srcBounds.Min.X == 0 && srcBounds.Min.Y == 0 {
		if src0, ok := img.(*image.NRGBA); ok {
			return src0
		}
	}
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y

	dstBounds := srcBounds.Sub(srcBounds.Min)
	dstW := dstBounds.Dx()
	dstH := dstBounds.Dy()
	dst := image.NewNRGBA(dstBounds)

	switch src := img.(type) {
	case *image.NRGBA:
		rowSize := srcBounds.Dx() * 4
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			si := src.PixOffset(srcMinX, srcMinY+dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				copy(dst.Pix[di:di+rowSize], src.Pix[si:si+rowSize])
			}
		}
	case *image.YCbCr:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				srcX := srcMinX + dstX
				srcY := srcMinY + dstY
				siy := src.YOffset(srcX, srcY)
				sic := src.COffset(srcX, srcY)
				r, g, b := color.YCbCrToRGB(src.Y[siy], src.Cb[sic], src.Cr[sic])
				dst.Pix[di+0] = r
				dst.Pix[di+1] = g
				dst.Pix[di+2] = b
				dst.Pix[di+3] = 0xff
				di += 4
			}
		}
	case *image.Gray:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			si := src.PixOffset(srcMinX, srcMinY+dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				c := src.Pix[si]
				dst.Pix[di+0] = c
				dst.Pix[di+1] = c
				dst.Pix[di+2] = c
				dst.Pix[di+3] = 0xff
				di += 4
				si += 2
			}
		}
	default:
		for dstY := 0; dstY < dstH; dstY++ {
			di := dst.PixOffset(0, dstY)
			for dstX := 0; dstX < dstW; dstX++ {
				c := color.NRGBAModel.Convert(img.At(srcMinX+dstX, srcMinY+dstY)).(color.NRGBA)
				dst.Pix[di+0] = c.R
				dst.Pix[di+1] = c.G
				dst.Pix[di+2] = c.B
				dst.Pix[di+3] = c.A
				di += 4
			}
		}
	}

	return dst
}
