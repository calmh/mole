var _ = require('underscore')._;
var childProcess = require('child_process');
var exec = childProcess.exec;
var spawn = childProcess.spawn;
var con = require('./console');

var interface;
if (process.platform === 'linux') {
    interface = 'lo';
} else if (process.platform === 'darwin' || process.platform === 'sunos') {
    interface = 'lo0';
} else {
    con.warning('Don\'t know the platform "' + process.platform + '", not setting up local IP:s for forwarding.');
}

module.exports = function (config, callback) {
    var missing = [];
    var ipMap = {};
    var ips;

    if (!interface) {
        return callback(false);
    }

    function addIp(ip, callback) {
        con.info('Adding local IP ' + ip + ' for forwards; if asked for password, give your local (sudo) password.');
        var ifconfig = spawn('sudo', ['ifconfig', interface, 'add', ip], { customFds: [0, 1, 2] });
        ifconfig.on('exit', callback);
    }

    function addMissingIps(exitCode) {
        if (exitCode === 0 && missing.length > 0) {
            addIp(missing.shift(), addMissingIps);
        } else {
            callback(exitCode === 0);
        }
    }

    _.each(config.forwards, function (fs, name) {
        fs.forEach(function (f) {
            var m = f.from.match(/^([0-9.]+):/);
            if (m) {
                ipMap[m[1]] = true;
            }
        });
    });
    ips = _.keys(ipMap);

    exec('ifconfig ' + interface, function (error, stdout, stderr) {
        ips.forEach(function (ip) {
            if (!stdout.match(new RegExp('\\s' + ip.replace('.', '\\.') + '\\s'))) {
                missing.push(ip);
            }
        });

        addMissingIps(0);
    });
}
