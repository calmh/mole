var fs = require('fs');
var should = require('should');

var vpncConfig = require('../lib/vpnc-config');

describe('vpncConfig', function () {
    it('should return the options', function () {
        var config = {
            vpnc: {
                foo_a: 'foo',
                bar_b: 'baz',
                baz_z: 'quux',
                quux_quux_quux: 'bar',
            }
        };

        var data = vpncConfig(config);
        data.should.match(/\bfoo a foo\n/);
        data.should.match(/\bbar b baz\n/);
        data.should.match(/\bbaz z quux\n/);
        data.should.match(/\bquux quux quux bar\n/);
    });
});
