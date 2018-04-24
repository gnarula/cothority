import { Point } from "../../index";
import { Ed25519 } from "./curve";
import BN from "bn.js";
import * as crypto from "crypto";
import Ed25519Scalar from "./scalar";

type PointType = number | Uint8Array | BN

export default class Ed25519Point implements Point {
    private curve: Ed25519
    private X: PointType
    private Y: PointType
    private Z: PointType
    private T: PointType
    private ref: {
        point: any,
        curve: Ed25519
    }

    constructor(curve: Ed25519, X?: PointType, Y?: PointType, Z?: PointType, T?: PointType) {
        let _X: PointType, _Y: PointType, _Z: PointType, _T: PointType;
        if (X !== undefined && X.constructor === Uint8Array) {
            _X = new BN(X, 16, "le");
        }
        if (Y !== undefined && Y.constructor === Uint8Array) {
            _Y = new BN(Y, 16, "le");
        }
        if (Z !== undefined && Z.constructor === Uint8Array) {
            _Z = new BN(Z, 16, "le");
        }
        if (T !== undefined && T.constructor === Uint8Array) {
            _T = new BN(T, 16, "le");
        }
        // the point reference is stored in a Point reference to make set()
        // consistent.
        this.ref = {
            point: curve.curve.point(_X, _Y, _Z, _T),
            curve: curve
        };
    }

    string(): string {
        return this.toString()
    }

    inspect(): string {
        return this.toString()
    }

    toString(): string {
        const bytes = this.marshalBinary();
        return Array.from(bytes, b =>
        ("0" + (b & 0xff).toString(16)).slice(-2)
        ).join("");
    }

    equal(p2: Ed25519Point): boolean {
        const b1 = this.marshalBinary();
        const b2 = p2.marshalBinary();
        for (var i = 0; i < 32; i++) {
            if (b1[i] !== b2[i]) {
                return false;
            }
        }
        return true;
    }

    null(): Ed25519Point {
        this.ref.point = this.ref.curve.curve.point(0, 1, 1, 0);
        return this;
    }

    base(): Ed25519Point {
        this.ref.point = this.ref.curve.curve.point(
            this.ref.curve.curve.g.getX(),
            this.ref.curve.curve.g.getY()
        );
        return this;
    }

    pick(callback?: (length: number) => Uint8Array): Ed25519Point {
        return this.embed(new Uint8Array([]), callback);
    }

    set(p: Ed25519Point): Ed25519Point {
        this.ref = p.ref;
        return this;
    }

    clone(): Ed25519Point {
        const { point } = this.ref;
        return new Ed25519Point(this.ref.curve, point.x, point.y, point.z, point.t);
    }

    embedLen(): number {
        // Reserve the most-significant 8 bits for pseudo-randomness.
        // Reserve the least-significant 8 bits for embedded data length.
        // (Hopefully it's unlikely we'll need >=2048-bit curves soon.)
        return Math.floor((255 - 8 - 8) / 8);
    }

    embed(data: Uint8Array, callback?: (length: number) => Uint8Array): Ed25519Point {
        let dl = this.embedLen();
        if (data.length > dl) {
            throw new Error("data.length > embedLen");
        }

        if (dl > data.length) {
            dl = data.length;
        }

        callback = callback || crypto.randomBytes;

        let point_obj = new Ed25519Point(this.ref.curve);
        while (true) {
            let buff = callback(32);
            let bytes = Uint8Array.from(buff);

            if (dl > 0) {
                bytes[0] = dl; // encode length in lower 8 bits
                bytes.set(data, 1); // copy in data to embed
            }

            let bnp = new BN(bytes, 16, "le");

            //if (bnp.cmp(PFSCALAR) > 0) {
            //continue; // try again
            //}

            try {
                point_obj.unmarshalBinary(bytes);
            } catch (e) {
                continue; // try again
            }
            if (dl == 0) {
                point_obj.ref.point = point_obj.ref.point.mul(new BN(8));
                if (point_obj.ref.point.isInfinity()) {
                    continue; // unlucky
                }
                return point_obj;
            }

            let q = point_obj.clone();
            q.ref.point = q.ref.point.mul(this.ref.curve.curve.n);
            if (q.ref.point.isInfinity()) {
                return point_obj;
            }
        }
    }

    data(): Uint8Array {
        const bytes = this.marshalBinary();
        const dl = bytes[0];
        if (dl > this.embedLen()) {
            throw new Error("invalid embedded data length");
        }
        return bytes.slice(1, dl + 1);
    }

    add(p1: Ed25519Point, p2: Ed25519Point): Ed25519Point {
        const point = p1.ref.point;
        this.ref.point = this.ref.curve.curve
            .point(point.x, point.y, point.z, point.t)
            .add(p2.ref.point);
        return this;
    }

    sub(p1: Ed25519Point, p2: Ed25519Point): Ed25519Point {
        const point = p1.ref.point;
        this.ref.point = this.ref.curve.curve
            .point(point.x, point.y, point.z, point.t)
            .add(p2.ref.point.neg());
        return this;
    }

    neg(p: Ed25519Point): Ed25519Point {
        this.ref.point = p.ref.point.neg();
        return this;
    }

    mul(s: Ed25519Scalar, p?: Ed25519Point): Ed25519Point {
        p = p || null;
        const arr = s.ref.arr.fromRed();
        this.ref.point =
        p !== null ? p.ref.point.mul(arr) : this.ref.curve.curve.g.mul(arr);
        return this;
    }

    marshalBinary(): Uint8Array {
        this.ref.point.normalize();

        const buffer = this.ref.point.getY().toArray("le", 32);
        buffer[31] ^= (this.ref.point.x.isOdd() ? 1 : 0) << 7;

        return Uint8Array.from(buffer);
    }

    unmarshalBinary(bytes: Uint8Array): void {
        if (bytes.constructor !== Uint8Array) {
        throw new TypeError("bytes should be a Uint8Array");
        }
        // we create a copy bcurveause the array might be modified
        const _bytes = new Uint8Array(32);
        _bytes.set(bytes, 0);

        const odd = _bytes[31] >> 7 === 1;

        _bytes[31] &= 0x7f;
        let bnp = new BN(_bytes, 16, "le");
        if (bnp.cmp(this.ref.curve.curve.p) >= 0) {
        throw new Error("bytes > p");
        }
        this.ref.point = this.ref.curve.curve.pointFromY(bnp, odd);
    }
}