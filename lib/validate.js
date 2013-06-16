"use strict";

var _ = require('underscore');

module.exports = function (config) {
    ['author', 'description'].forEach(function (attr) {
        if (!config.general[attr]) {
            throw new Error('Missing required attribute "' + attr + '"');
        }
    });

    ['hosts'].forEach(function (attr) {
        if (!config[attr]) {
            throw new Error('Missing required attribute "' + attr + '"');
        }
    });

    _.each(config.hosts, function (host, name) {
        ['addr', 'user'].forEach(function (attr) {
            if (!host[attr]) {
                throw new Error('Missing required attribute "' + attr + '" on host "' + name + '"');
            }
        });

        if (!host.password && !host.key) {
            throw new Error('Missing required attribute "password" or "key" on host "' + name + '"');
        }

        if (host.socks && host.via) {
            throw new Error('Cannot use "via" and "socks" on same host "' + name + '"');
        }
    });

    if (!config.general.main && _.size(config.forwards) === 0) {
        throw new Error('Missing either "main" or "forward" directives.');
    }

    if (config.general.main && !config.hosts[config.general.main]) {
        throw new Error('Missing main host "' + config.general.main + '"');
    }

    if (config.general.aliases) {
        config.general.aliases.forEach(function (alias) {
            var l = alias.trim().split(/\s+/);
            if (l.length !== 2) {
                throw new Error('Malformed alias: ' + alias);
            }
            if (!l[0].match(/^[0-9a-z-_.]+$/)) {
                throw new Error('Malformed alias name: ' + l[0]);
            }
            if (!l[1].match(/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+$/)) {
                throw new Error('Malformed alias IP: ' + l[1]);
            }
        });
    }

    var seen = {};
    _.values(config.forwards).forEach(function (fwd) {
        _.each(fwd, function (to, from) {
            if (!from.match(/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$/)) {
                throw new Error('Malformed forward from: ' + from);
            }
            if (!to.match(/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$/)) {
                throw new Error('Malformed forward to: ' + to);
            }
            if (seen[from]) {
                throw new Error('Duplicate forward from: ' + from);
            }
            seen[from] = true;
        })
    });

    var validAttrs = ['general', 'hosts', 'forwards', 'vpnc', 'vpn routes', 'openconnect'];
    _.each(config, function (val, attr) {
        if (!_.contains(validAttrs, attr)) {
            throw new Error('Unknown config attribute "' + attr + '"');
        }
    });

    return true;
};

