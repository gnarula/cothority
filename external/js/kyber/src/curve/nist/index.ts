import NistPoint from "./point";
import NistScalar from "./scalar";
import Weierstrass from "./curve";
import Params from "./params"

export default {
    Point: NistPoint,
    Scalar: NistScalar,
    Curve: Weierstrass,
    Params,
}