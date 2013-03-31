"use strict";

var _ = require('underscore');
var fs = require('fs');
var ini = require('ini');
var path = require('path');

var validate = require('./validate');
var con = require('./console');

module.exports = {
    name: name,
    loadByName: loadByName,
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

function loadByName(name, tunnelDefDir) {
    var local = path.join(tunnelDefDir, name) + '.ini';
    return loadFile(local);
}

function loadFile(local) {
    if (!fs.existsSync(local)) {
        throw new Error('File "' + local + '" does not exist');
    }

    var obj = loadIniTunnel(local);
    obj.stat = fs.statSync(local);

    validate(obj);
    return obj;
}

function loadIniTunnel(name) {
    var str = fs.readFileSync(name, 'utf-8');
    return parseIniTunnel(str);
}

function parseIniTunnel(str) {
    var obj = ini.parse(str);
    if (obj.general.version === '3') {
        if (!obj.hosts)
            obj.hosts = {};
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

function save(config, name) {
    config.general.version = 3;
    fs.writeFileSync(name, ini.stringify(config));
}
