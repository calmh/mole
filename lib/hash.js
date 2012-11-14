"use strict";

var crypto = require('crypto');
var fs = require('fs');

exports = module.exports = function hash(file, callback) {
    fs.readFile(file, function(err, data) {
        if (err) return callback(err);

        var hash = crypto.createHash('sha1');
        hash.update(data);
        var sha1 = hash.digest('hex');
        callback(null, sha1);
    });
};

exports.sync = function hashSync(file) {
    var data = fs.readFileSync(file);
    var hash = crypto.createHash('sha1');
    hash.update(data);
    return hash.digest('hex');
};
