package bmpfonts

type AtlasDescription struct {
    IndexOfCapitalA int
    IndexOfSmallA   *int
    IndexOfZero     *int
    Chains          []SpecialCharacterChain
}
type SpecialCharacterChain struct {
    StartIndex int
    Characters []rune
}

func NewIndexFromDescription(desc AtlasDescription) map[rune]uint16 {
    result := map[rune]uint16{}

    indexOfCapitalA := desc.IndexOfCapitalA

    indexOfZero := indexOfCapitalA + 26
    if desc.IndexOfZero != nil {
        indexOfZero = *desc.IndexOfZero
    }

    for i := 0; i < 26; i++ {
        result[rune(i+65)] = uint16(indexOfCapitalA + i)
    }

    for i := 0; i < 10; i++ {
        result[rune(i+48)] = uint16(indexOfZero + i)
    }

    if desc.IndexOfSmallA != nil {
        indexOfSmallA := *desc.IndexOfSmallA
        for i := 0; i < 26; i++ {
            result[rune(i+97)] = uint16(indexOfSmallA + i)
        }
    }

    for _, chain := range desc.Chains {
        indexOfSpecialChain := chain.StartIndex
        for i, special := range chain.Characters {
            result[special] = uint16(indexOfSpecialChain + i)
        }
    }

    return result
}
