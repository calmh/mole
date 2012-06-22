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

    if (!config.main && _.size(config.localForwards) === 0) {
        throw new Error('Missing either "main" or "localforward" directives.');
    }

    if (config.main && !config.hosts[config.main]) {
        throw new Error('Missing main host "' + config.main + '"');
    }

    return true;
};

