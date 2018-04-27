import * as curve from "./curve";
import sign from "./sign";

export interface Group {
  scalarLen(): number;

  scalar(): Scalar;

  pointLen(): number;

  point(): Point;
}

export interface Point {
  equal(p2: Point): boolean;

  null(): Point;

  base(): Point;

  pick(callback?: (length: number) => Uint8Array): Point;

  set(p: Point): Point;

  clone(): Point;

  embedLen(): number;

  embed(data: Uint8Array, callback?: (length: number) => Uint8Array): Point;

  data(): Uint8Array;

  add(p1: Point, p2: Point): Point;

  sub(p1: Point, p2: Point): Point;

  neg(p: Point): Point;

  mul(s: Scalar, p?: Point): Point;

  marshalBinary(): Uint8Array;

  unmarshalBinary(bytes: Uint8Array): void;
}

export interface Scalar {
  marshalBinary(): Uint8Array;

  unmarshalBinary(bytes: Uint8Array): void;

  equal(s2: Scalar): boolean;

  set(a: Scalar): Scalar;

  clone(): Scalar;

  zero(): Scalar;

  add(a: Scalar, b: Scalar): Scalar;

  sub(a: Scalar, b: Scalar): Scalar;

  neg(a: Scalar): Scalar;

  div(a: Scalar, b: Scalar): Scalar;

  mul(s1: Scalar, b: Scalar): Scalar;

  inv(a: Scalar): Scalar;

  one(): Scalar;

  pick(callback?: (length: number) => Uint8Array): Scalar;

  setBytes(bytes: Uint8Array): Scalar;
}

export default {
  curve,
  sign
}