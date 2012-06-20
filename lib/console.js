"use strict";

var colors = require('colors');
var debug = false;

exports.enableDebug = function () {
    debug = true;
};

exports.debug = function (str) {
    if (debug) {
        console.log('debug'.magenta.bold + ' - ' + str);
    }
};

exports.fatal = function (str) {
    console.log('fatal'.red.bold + ' - ' + str);
    process.exit(-1);
};

exports.error = function(str) {
    console.log('error'.red.bold + ' - ' + str);
};

exports.warning = function(str) {
    console.log('warning'.yellow.bold + ' - ' + str);
};

exports.ok = function(str) {
    console.log('ok'.green.bold + ' - ' + str);
};

exports.info = function(str) {
    console.log('info'.blue.bold + ' - ' + str);
};
