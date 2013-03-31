var should = require('should');
var fs = require('fs');

var users = require('../lib/jsonstore');
var testFile;
var cnt = 0;

describe('users', function () {
    beforeEach(function () {
        try {
            cnt += 1;
            testFile = '/tmp/users.json' + cnt;
            fs.unlinkSync(testFile);
        } catch (err) {
            // Never mind.
        }
    });

    it('should get and set users', function () {
        var u = new users();
        u.set('test', { a:1, b:2 });

        u.get('test').should.eql({ a:1, b:2 });
    });

    it('should save and load users', function (done) {
        var u = new users(testFile);
        u.set('test', { a:1, b:2 });
        u.saveNow(function () {
            var t = new users(testFile);
            t.get('test').should.eql({ a:1, b:2 });
            done();
        });
    });

    it('should save users automatically', function (done) {
        var u = new users(testFile);
        u.set('test', { a:1, b:2 });
        u.set('test2', { a:2, b:3 });

        setTimeout(function () {
            var t = new users(testFile);
            t.get('test').should.eql({ a:1, b:2 });
            t.get('test2').should.eql({ a:2, b:3 });
            done();
        }, 300);
    });

    it('should return all users', function () {
        var u = new users(testFile);
        u.set('test', { a:1, b:2 });
        u.set('test2', { a:2, b:3 });
        u.all().length.should.equal(2);
    });

    it('should delete users', function () {
        var u = new users(testFile);
        u.set('test', { a:1, b:2 });
        u.set('test2', { a:2, b:3 });
        u.del('test');
        u.all().length.should.equal(1);
        should.not.exist(u.get('test'));
    });
});
