"use strict";

var con = require('yacon');
var exec = require('child_process').exec;
var mkdirp = require('mkdirp');
var os = require('os');
var path = require('path');
var sudo = require('sudo');
var temp = require('temp');


module.exports = install;
install.help = 'Install an optional package, fetched from the server';
install.options = {
    pkg: { position: 1, help: 'Package name', required: true }
};
install.prio = 5;

function install(opts, state) {
    // We build the expected package name based on the name specified by the
    // user, plus the platform and architecture.

    var file = [ opts.pkg, os.platform(), os.arch() ].join('-') + '.tar.gz';
    var local = path.join(state.path.pkgDir, file);

    // Get the package from the server and save it in our package directory.
    // The callback will be called only if the fetch and save is successfull.

    con.info('Fetching ' + file);
    state.client.saveBin('/extra/' + file, local, function () {

        // Create a temporary path where we can extract the package.

        var tmp = temp.path();
        mkdirp(tmp);

        // Change working directory to the temporary one we created and try to
        // extract the downloaded package file.

        con.info('Unpacking ' + file);
        exec('cd ' + tmp + ' && tar zxf ' + local, function (err, stdout, stderr) {
            con.debug('Extracted in ' + tmp);

            // The package should include a script `install.sh` that will do
            // whatever's necessary to install the package. We run that with
            // sudo.

            con.info('Running installation, you might now be asked for your local (sudo) password.');
            var inst = sudo([path.join(tmp, 'install.sh'), tmp]);
            inst.on('exit', function (code) {

                // We're done, one way or the other.

                if (code === 0) {
                    con.ok('Installation complete');
                } else {
                    con.info('Installation failed. Sorry.');
                }
            });
        });
    });
}
