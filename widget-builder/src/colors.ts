function parseColor(input: string): [number, number, number] | null {
  const s = input.trim().toLowerCase()
  const hex = s.match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/)
  if (hex) {
    const h = hex[1].length === 3 ? hex[1].replace(/./g, (c) => c + c) : hex[1]
    return [parseInt(h.slice(0, 2), 16), parseInt(h.slice(2, 4), 16), parseInt(h.slice(4, 6), 16)]
  }
  const rgb = s.match(/^rgba?\(\s*(\d+)[,\s]+(\d+)[,\s]+(\d+)/)
  return rgb ? [Number(rgb[1]), Number(rgb[2]), Number(rgb[3])] : null
}

function luminance([r, g, b]: [number, number, number]): number {
  const [R, G, B] = [r, g, b].map((v) => {
    const c = v / 255
    return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4
  })
  return 0.2126 * R + 0.7152 * G + 0.0722 * B
}

/** WCAG contrast ratio between two CSS colors (hex or rgb()); null if either can't be parsed. */
export function contrastRatio(a: string, b: string): number | null {
  const ca = parseColor(a)
  const cb = parseColor(b)
  if (!ca || !cb) return null
  const [hi, lo] = [luminance(ca), luminance(cb)].sort((x, y) => y - x)
  return (hi + 0.05) / (lo + 0.05)
}

/** Mantine's default palette shade 6 for the named primary colors the widget accepts. */
export const NAMED_COLORS: Record<string, string> = {
  dark: '#2e2e2e', gray: '#868e96', red: '#fa5252', pink: '#e64980', grape: '#be4bdb',
  violet: '#7950f2', indigo: '#4c6ef5', blue: '#228be6', cyan: '#15aabf', green: '#40c057',
  lime: '#82c91e', yellow: '#fab005', orange: '#fd7e14', teal: '#12b886',
}
