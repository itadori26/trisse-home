/**
 * Arithmetic for the smart search (#160): a tokenizer and a recursive
 * descent parser that evaluates while it parses (interpreter pattern).
 * Pure functions; invalid or ambiguous input yields null, never an error,
 * because the search evaluates every keystroke.
 *
 * Grammar (lowest precedence first):
 *   sum     = product (("+" | "-") product)*      a + b % adds b percent of a
 *   product = unary (("*" | "/") unary)*
 *   unary   = "-" unary | "+" unary | power        -2^2 = -4
 *   power   = postfix ("^" unary)?                right associative
 *   postfix = primary ("%" (("von" | "of") unary)?)?
 *   primary = number | "(" sum ")"
 */

type Token =
  | { kind: 'num'; value: number }
  | { kind: 'op'; op: '+' | '-' | '*' | '/' | '^' | '%' | '(' | ')' }
  | { kind: 'of' };

const OPS: Record<string, '+' | '-' | '*' | '/' | '^' | '%' | '(' | ')'> = {
  '+': '+',
  '-': '-',
  '−': '-',
  '*': '*',
  '×': '*',
  '·': '*',
  x: '*',
  '/': '/',
  ':': '/',
  '÷': '/',
  '^': '^',
  '%': '%',
  '(': '(',
  ')': ')',
};

const OF_WORDS = new Set(['von', 'of']);

/**
 * Reads a number with decimal comma or point. With both separators the
 * last one is the decimal mark ("1.234,5", "1,234.5"); one kind used
 * several times groups thousands ("1.000.000"). Groups must have three
 * digits, so "13.11.2026" is no number. Integers with a leading zero
 * ("0151") look like phone numbers or ids and are rejected.
 */
export function parseNumber(text: string): number | null {
  if (!/^\d[\d.,]*$/.test(text) || /[.,]$/.test(text)) return null;
  const dots = (text.match(/\./g) ?? []).length;
  const commas = (text.match(/,/g) ?? []).length;
  let normalised: string;
  if (dots > 0 && commas > 0) {
    const decimal = text.lastIndexOf('.') > text.lastIndexOf(',') ? '.' : ',';
    const group = decimal === '.' ? ',' : '.';
    if ((decimal === '.' ? dots : commas) !== 1) return null;
    const [int = '', frac = ''] = text.split(decimal);
    if (!/^\d{1,3}(\d{3})*$/.test(int.replaceAll(group, '')) || !groupedOk(int, group)) return null;
    normalised = `${int.replaceAll(group, '')}.${frac}`;
  } else if (dots + commas > 1) {
    const group = dots > 0 ? '.' : ',';
    if (!groupedOk(text, group)) return null;
    normalised = text.replaceAll(group, '');
  } else normalised = text.replace(',', '.');
  if (/^0\d/.test(normalised)) return null;
  const n = Number(normalised);
  return Number.isFinite(n) ? n : null;
}

/** "1.234.567": first group 1–3 digits, every further group exactly three. */
function groupedOk(text: string, group: string): boolean {
  const parts = text.split(group);
  return /^\d{1,3}$/.test(parts[0] ?? '') && parts.slice(1).every((p) => /^\d{3}$/.test(p));
}

/** Splits an expression into tokens; null on any unknown character. */
export function tokenizeExpression(input: string): Token[] | null {
  const tokens: Token[] = [];
  const re = /\s*(?:(\d[\d.,]*)|([a-zäöü]+)|(\S))/giy;
  const text = input.trim();
  let m: RegExpExecArray | null;
  while (re.lastIndex < text.length && (m = re.exec(text))) {
    const [, num, word, sym] = m;
    if (num !== undefined) {
      const value = parseNumber(num);
      if (value === null) return null;
      tokens.push({ kind: 'num', value });
    } else if (word !== undefined) {
      const w = word.toLowerCase();
      if (OF_WORDS.has(w)) tokens.push({ kind: 'of' });
      else if (w === 'x') tokens.push({ kind: 'op', op: '*' });
      else return null;
    } else if (sym !== undefined) {
      const op = OPS[sym];
      if (!op) return null;
      tokens.push({ kind: 'op', op });
    }
  }
  return re.lastIndex >= text.length || text.length === 0 ? tokens : null;
}

class Fail extends Error {}

/** A value and whether it is a bare percentage ("10 %"), for a + b %. */
interface Value {
  n: number;
  percent: boolean;
}

class Parser {
  private i = 0;
  constructor(private readonly tokens: Token[]) {}

  parse(): number {
    const v = this.sum();
    if (this.i !== this.tokens.length) throw new Fail();
    return v.n;
  }

  private peekOp(): string | null {
    const t = this.tokens[this.i];
    return t?.kind === 'op' ? t.op : null;
  }

  private sum(): Value {
    let left = this.product();
    for (let op = this.peekOp(); op === '+' || op === '-'; op = this.peekOp()) {
      this.i++;
      const right = this.product();
      // "200 + 10 %" is 220 like on a calculator, "200 - 10 %" is 180.
      const delta = right.percent ? left.n * right.n : right.n;
      left = { n: op === '+' ? left.n + delta : left.n - delta, percent: false };
    }
    return left;
  }

  private product(): Value {
    let left = this.unary();
    for (let op = this.peekOp(); op === '*' || op === '/'; op = this.peekOp()) {
      this.i++;
      const right = this.unary();
      if (op === '/' && right.n === 0) throw new Fail();
      left = { n: op === '*' ? left.n * right.n : left.n / right.n, percent: false };
    }
    return left;
  }

  private unary(): Value {
    const op = this.peekOp();
    if (op === '-' || op === '+') {
      this.i++;
      const v = this.unary();
      return { n: op === '-' ? -v.n : v.n, percent: v.percent };
    }
    return this.power();
  }

  private power(): Value {
    const base = this.postfix();
    if (this.peekOp() !== '^') return base;
    this.i++;
    const exp = this.unary();
    return { n: base.n ** exp.n, percent: false };
  }

  private postfix(): Value {
    const v = this.primary();
    if (this.peekOp() !== '%') return v;
    this.i++;
    if (this.tokens[this.i]?.kind === 'of') {
      this.i++;
      // "19 % von 240" = 45.6
      return { n: (v.n / 100) * this.unary().n, percent: false };
    }
    return { n: v.n / 100, percent: true };
  }

  private primary(): Value {
    const t = this.tokens[this.i++];
    if (t?.kind === 'num') return { n: t.value, percent: false };
    if (t?.kind === 'op' && t.op === '(') {
      const v = this.sum();
      if (this.peekOp() !== ')') throw new Fail();
      this.i++;
      return { n: v.n, percent: false };
    }
    throw new Fail();
  }
}

/** Evaluates tokens; null when they are no complete, finite expression. */
export function evaluateTokens(tokens: Token[]): number | null {
  if (tokens.length === 0) return null;
  try {
    const n = new Parser(tokens).parse();
    return Number.isFinite(n) ? n : null;
  } catch (err) {
    if (err instanceof Fail) return null;
    throw err;
  }
}

/** Dates and phone numbers look like arithmetic but are none. */
const LOOKS_LIKE_DATA = [
  /^\d{4}-\d{1,2}-\d{1,2}$/, // 2026-11-13
  /^\d{1,2}\/\d{1,2}\/\d{2,4}$/, // 11/13/2026
  /^(?:\+|0)\d*(?:[\s-]\d+)+$/, // 0151 123 456, +49-151-1234 (but 10-4-3 is arithmetic)
];

/**
 * Evaluates a typed arithmetic expression. Answers only when there is
 * something to calculate: a bare number ("42") or a date gives null.
 */
export function evaluate(input: string): number | null {
  const text = input.trim();
  if (LOOKS_LIKE_DATA.some((re) => re.test(text))) return null;
  const tokens = tokenizeExpression(text);
  if (!tokens) return null;
  // "-5" or "50 %" alone is a number, not a calculation.
  if (tokens.filter((t) => t.kind === 'num').length < 2) return null;
  return evaluateTokens(tokens);
}
