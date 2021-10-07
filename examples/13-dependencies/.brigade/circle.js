// Approach gleaned from https://stackoverflow.com/a/4483755/3084239

// PI will not be accessible from outside this module
var PI = 3.14;

exports.area = function (r) {
    return PI * r * r;
};

exports.circumference = function (r) {
    return 2 * PI * r;
};