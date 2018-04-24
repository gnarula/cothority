import { Scalar, Group } from "../../index";
import BN, { Red } from "bn.js";
import * as crypto from "crypto";
import { Ed25519 } from "./curve";
import { int } from "../../random"

export default class Ed25519Scalar implements Scalar {
    ref: {
        arr: BN,
        curve: Ed25519
        red: Red
    }

    constructor(curve: Ed25519, red: Red) {
        this.ref = {
            arr: new BN(0, 16).toRed(red),
            curve: curve,
            red: Red
        };
    }

    marshalSize(): number {
        return 32;
    }

    marshalBinary(): Uint8Array {
        return new Uint8Array(this.ref.arr.fromRed().toArray("le", 32));
    }

    unmarshalBinary(bytes: Uint8Array): void {
        if (bytes.length > this.marshalSize()) {
            throw new Error("bytes.length > marshalSize");
        }
        this.ref.arr = new BN(bytes, 16, "le").toRed(this.ref.red);
    }

    equal(s2: Ed25519Scalar): boolean {
        return this.ref.arr.fromRed().cmp(s2.ref.arr.fromRed()) == 0;
    }

    set(a: Ed25519Scalar): Ed25519Scalar {
        this.ref = a.ref;
        return this;
    }

    clone(): Scalar {
        return new Ed25519Scalar(this.ref.curve, this.ref.red).setBytes(
            new Uint8Array(this.ref.arr.fromRed().toArray("le"))
        );
    }

    zero(): Scalar {
        this.ref.arr = new BN(0, 16).toRed(this.ref.red);
        return this;
    }
    add(a: Ed25519Scalar, b: Ed25519Scalar): Ed25519Scalar {
        this.ref.arr = a.ref.arr.redAdd(b.ref.arr);
        return this;
    }

    sub(a: Ed25519Scalar, b: Ed25519Scalar): Ed25519Scalar {
        this.ref.arr = a.ref.arr.redSub(b.ref.arr);
        return this;
    }

    neg(a: Ed25519Scalar): Ed25519Scalar {
        this.ref.arr = a.ref.arr.redNeg();
        return this;
    }

    mul(s1: Ed25519Scalar, s2: Ed25519Scalar): Ed25519Scalar {
        this.ref.arr = s1.ref.arr.redMul(s2.ref.arr);
        return this;
    }

    div(s1: Ed25519Scalar, s2: Ed25519Scalar): Ed25519Scalar {
        this.ref.arr = s1.ref.arr.redMul(s2.ref.arr.redInvm());
        return this;
    }

    inv(a: Ed25519Scalar): Ed25519Scalar {
        this.ref.arr = a.ref.arr.redInvm();
        return this;
    }

    one(): Ed25519Scalar {
        this.ref.arr = new BN(1, 16).toRed(this.ref.red);
        return this;
    }
    pick(callback?: (length: number) => Uint8Array): Scalar {
        callback = callback || crypto.randomBytes;
        const bytes = int(this.ref.curve.curve.n, callback);
        this.ref.arr = new BN(bytes, 16).toRed(this.ref.red);
        return this;
    }

    setBytes(bytes: Uint8Array): Scalar {
        this.ref.arr = new BN(bytes , 16, "le").toRed(this.ref.red);
        return this;
    }
}