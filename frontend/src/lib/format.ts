const PLAIN_DECIMAL = /^(-?)(\d+)(\.\d+)?$/;
const THOUSANDS = /\B(?=(\d{3})+$)/g;

// Grouped on the string itself: Intl.NumberFormat turns integers past ~1e308 into "∞", and
// results can have 1000 digits. The fraction is kept verbatim.
export function formatDecimal(value: string): string {
  const match = PLAIN_DECIMAL.exec(value);
  if (!match) return value;

  const [, sign, integer, fraction = ""] = match;
  return sign + integer.replace(THOUSANDS, ",") + fraction;
}
