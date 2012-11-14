"use strict";

/*global it: false, describe: false */

var fs = require('fs');
var hash = require('../lib/hash');
var should = require('should');

describe('validate', function () {
    it('should handle trivial data', function () {
        var data = 'abc123\n';
        var correct = '61ee8b5601a84d5154387578466c8998848ba089'; // echo abc123 | shasum
        fs.writeFileSync('/tmp/hashtest', data, 'utf8');
        var result = hash.sync('/tmp/hashtest');
        result.should.equal(correct);
    });

    it('should handle trivial data (async)', function (done) {
        var data = 'abc123\n';
        var correct = '61ee8b5601a84d5154387578466c8998848ba089';
        fs.writeFileSync('/tmp/hashtest', data, 'utf8');
        var result = hash('/tmp/hashtest', function (err, result) {
            result.should.equal(correct);
            done();
        });
    });

    it('should handle utf8 data', function () {
        var data = 'abcåäö\n';
        var correct = 'ef0105b202a329a7977599884c8ac9a728b30bae'; // echo abcåäö | shasum
        fs.writeFileSync('/tmp/hashtest', data, 'utf8');
        var result = hash.sync('/tmp/hashtest');
        result.should.equal(correct);
    });

    it('should handle utf8 data (async)', function (done) {
        var data = 'abcåäö\n';
        var correct = 'ef0105b202a329a7977599884c8ac9a728b30bae';
        fs.writeFileSync('/tmp/hashtest', data, 'utf8');
        hash('/tmp/hashtest', function (err, result) {
            result.should.equal(correct);
            done();
        });
    });
});
