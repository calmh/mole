"use strict";

var _ = require('underscore');
var con = require('yacon');
var fs = require('fs');
var ini = require('ini');
var path = require('path');

var ranges = require('./ranges');
var validate = require('./validate');

module.exports = {
    name: name,
    loadFile: loadFile,
    save: save,
    clean: clean,
    parse: parseIniTunnel
};

function clean(str) {
    return ini.stringify(ini.parse(str));
}

function name(file) {
    return path.basename(file).replace(/\.ini$/, '');
}

function loadFile(local) {
    if (!fs.existsSync(local)) {
        throw new Error('File "' + local + '" does not exist');
    }

    var str = fs.readFileSync(local, 'utf-8');
    var obj = parseIniTunnel(str);

    validate(obj);
    return obj;
}

function parseIniTunnel(str) {
    var obj = ini.parse(str);
    if (obj.general.version === '3') {
        if (!obj.hosts)
            obj.hosts = {};

        if (obj.forwards) {
            for (var fwdName in obj.forwards) {
                obj.forwards[fwdName] = expandForwardRanges(obj.forwards[fwdName]);
            }
        }
        return obj;
    }

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

function expandForwardRanges(fwd) {
    for (var from in fwd) {
        var to = fwd[from];
        var fromParts = from.split(':');
        if (to.indexOf(':') === -1) {
            delete fwd[from];
            ranges.expand(fromParts[1]).forEach(function (port) {
                fwd[fromParts[0] + ':' + port] = to + ':' + port;
            })
        }
    }
    return fwd;
}

function save(config, name) {
    config.general.version = 3;
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
    fs.writeFileSync(name, ini.stringify(config));
}
