"use strict";

var _ = require('underscore');
var fs = require('fs');
var ini = require('ini');
var path = require('path');

var ranges = require('./ranges');
var validate = require('./validate');

var formatVersion = 3.2;

module.exports = {
    load: load,
    save: save,
    parse: parse,
    formatVersion: formatVersion
};

function load(local) {
    if (!fs.existsSync(local)) {
        throw new Error('File "' + local + '" does not exist');
    }

    var str = fs.readFileSync(local, 'utf-8');
    var obj = parse(str);

    validate(obj);
    return obj;
}

function parse(str) {
    var obj = ini.parse(str);
    if (obj.general.version) {
        obj.general.version = parseFloat(obj.general.version);
    } else {
        obj.general.version = 2.0;
    }

    if (obj.general.version > formatVersion) {
        // Maximum supported version
        throw new Error('Config format version ' + obj.general.version + ' is not supported.');
    }

    Object.defineProperty(obj, '$', {value: {}, enumerate: false});

    if (obj.general.version >= 3.2) {
        // Support hidden optional attributes in general
        processForwardComments(obj);
    }

    if (obj.general.version >= 3.0) {
        if (!obj.hosts)
            obj.hosts = {};

        if (obj.forwards) {
            for (var fwdName in obj.forwards) {
                obj.forwards[fwdName] = expandForwardRanges(obj.forwards[fwdName]);
            }
        }
        return obj;
    }

    // Version 2 and earlier
    var config = {};
    config.general = obj.general;
    config.hosts = {};
    config.forwards = {};

    _.each(obj, function (val, key) {
        var m, arr;

        // Host sections look like [host host_name]
        m = key.match(/^host ([^ ]+)$/);
        if (m) {
            // SSH keys have newlines replaced by spaces
            if (val.key) {
                val.key = val.key.replace(/ /g, '\n').replace(/\nRSA\nPRIVATE\nKEY/g, ' RSA PRIVATE KEY');
            }
            config.hosts[m[1]] = val;
            return;
        }

        // Forward sections look like [forward A description here]
        m = key.match(/^(?:local)?forward +(.+)$/);
        if (m) {
            config.forwards[m[1]] = val;
            return;
        }

        if (key !== 'general') {
            config[key] = val;
            return;
        }
    });
    return config;
}

function processForwardComments(obj) {
    obj.$.forwardComments = {};
    for (var fwdName in obj.forwards) {
        for (var from in obj.forwards[fwdName]) {
            if (from === 'comment') {
                obj.$.forwardComments[fwdName] = obj.forwards[fwdName].comment;
                delete obj.forwards[fwdName].comment;
            }
        }
    }
}

function expandForwardRanges(fwd) {
    var res = {};
    for (var from in fwd) {
        var to = fwd[from];
        var fromParts = from.split(':');
        if (fromParts.length != 2) {
            throw new Error('Malformed forward from in "' + from + '".')
        }
        if (to.indexOf(':') === -1) {
            ranges.expand(fromParts[1]).forEach(function (port) {
                res[fromParts[0] + ':' + port] = to + ':' + port;
            })
        } else {
            res[from] = to;
        }
    }
    return res;
}

function save(config, name) {
    if (config.general.version < 3.0) {
        config.general.version = 3.0;
    }

    _.each(config.forwards, function (fwd, desc) {
        var grouped = {};
        var newFwd = {};
        var keys = Object.keys(fwd);
        keys.sort();

        for (var i = 0; i < keys.length; i++) {
            var fromParts = keys[i].split(':');
            var toParts = fwd[keys[i]].split(':');
            if (fromParts[1] !== toParts[1]) {
                // Not a candidate for aggregating
                newFwd[keys[i]] = fwd[keys[i]];
            } else {
                var g = fromParts[0] + '/' + toParts[0];
                if (!grouped[g])
                    grouped[g] = [];
                grouped[g].push(parseInt(toParts[1], 10));
            }
        }

        for (var g in grouped) {
            var parts = g.split('/');
            var from = parts[0];
            var to = parts[1];
            var ra = ranges.compress(grouped[g]);
            ra.forEach(function (r) {
                newFwd[from + ':' + r] = to;
            })
        }

        config.forwards[desc] = newFwd;
    });

    if (config.$.forwardComments) {
        if (config.general.version < 3.2) {
            config.general.version = 3.2;
        }

        for (var key in config.$.forwardComments) {
            config.forwards[key].comment = config.$.forwardComments[key];
        }
    }

    fs.writeFileSync(name, ini.stringify(config));
}
