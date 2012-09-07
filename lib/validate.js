"use strict";

var _ = require('underscore');

module.exports = function (config) {
    ['author', 'description', 'hosts'].forEach(function (attr) {
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
    });

    if (!config.main && _.size(config.forwards) === 0) {
        throw new Error('Missing either "main" or "forward" directives.');
    }

    if (config.main && !config.hosts[config.main]) {
        throw new Error('Missing main host "' + config.main + '"');
    }

    _.flatten(_.values(config.forwards)).forEach(function (fwd) {
        if (!fwd.from.match(/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$/)) {
            throw new Error('Malformed forward "from: ' + fwd.from + '"');
        }
        if (!fwd.to.match(/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+:[0-9]+$/)) {
            throw new Error('Malformed forward "to: ' + fwd.from + '"');
        }
    });

    // FIXME: stat should not be an allowed attribute. Solve it some other way.
    var validAttrs = ['author', 'description', 'main', 'hosts', 'forwards', 'vpnc', 'vpnRoutes', 'openconnect', 'stat'];
    _.each(config, function (val, attr) {
        if (!_.contains(validAttrs, attr)) {
            throw new Error('Unknown config attribute "' + attr + '"');
        }
    });

    return true;
};

