var colors = require('colors');
var debug = false;

exports.enableDebug = function () {
    debug = true;
}

exports.debug = function (str) {
    if (debug) {
        console.log('debug'.bold + ' - ' + str);
    }
}

exports.fatal = function (str) {
    console.log('fatal'.red.bold + ' - ' + str);
    process.exit(-1);
}

exports.warning = function(str) {
    console.log('warning'.yellow + ' - ' + str);
    process.exit(-1);
}

exports.ok = function(str) {
    console.log('ok'.green + ' - ' + str);
}

exports.info = function(str) {
    console.log('info'.bold + ' - ' + str);
}
