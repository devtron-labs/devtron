const {series, watch, src, dest, parallel} = require('gulp');
const pump = require('pump');

// gulp plugins and utils
const livereload = require('gulp-livereload');
const postcss = require('gulp-postcss');
const zip = require('gulp-zip');
const concat = require('gulp-concat');
const uglify = require('gulp-uglify');
const beeper = require('beeper');
const fs = require('fs');

// postcss plugins
const autoprefixer = require('autoprefixer');
const colorFunction = require('postcss-color-function');
const cssnano = require('cssnano');
const customProperties = require('postcss-custom-properties');
const easyimport = require('postcss-easy-import');

function serve(done) {
    livereload.listen();
    done();
}

const handleError = (done) => {
    return function (err) {
        if (err) {
            beeper();
        }
        return done(err);
    };
};

function hbs(done) {
    pump([
        src(['*.hbs', 'partials/**/*.hbs']),
        livereload()
    ], handleError(done));
}

function css(done) {
    const processors = [
        easyimport,
        customProperties({preserve: false}),
        colorFunction(),
        autoprefixer(),
        cssnano()
    ];

    pump([
        src('assets/css/*.css', {sourcemaps: true}),
        postcss(processors),
        dest('assets/built/', {sourcemaps: '.'}),
        livereload()
    ], handleError(done));
}

function js(done) {
    pump([
        src([
            // pull in lib files first so our own code can depend on it
            'assets/js/lib/*.js',
            'assets/js/*.js'
        ], {sourcemaps: true}),
        concat('casper.js'),
        uglify(),
        dest('assets/built/', {sourcemaps: '.'}),
        livereload()
    ], handleError(done));
}

function zipper(done) {
    const targetDir = 'dist/';
    const themeName = require('./package.json').name;
    const filename = themeName + '.zip';

    pump([
        src([
            '**',
            '!node_modules', '!node_modules/**',
            '!dist', '!dist/**'
        ]),
        zip(filename),
        dest(targetDir)
    ], handleError(done));
}

const cssWatcher = () => watch('assets/css/**', css);
const hbsWatcher = () => watch(['*.hbs', 'partials/**/*.hbs'], hbs);
const watcher = parallel(cssWatcher, hbsWatcher);
const build = series(css, js);
const dev = series(build, serve, watcher);

exports.build = build;
exports.zip = series(build, zipper);
exports.default = dev;

// release imports
const path = require('path');
const releaseUtils = require('@tryghost/release-utils');

let config;
try {
    config = require('./config');
} catch (err) {
    config = null;
}

const REPO = 'TryGhost/Casper';
const USER_AGENT = 'Casper';
const CHANGELOG_PATH = path.join(process.cwd(), '.', 'changelog.md');

const changelog = ({previousVersion}) => {
    const changelog = new releaseUtils.Changelog({
        changelogPath: CHANGELOG_PATH,
        folder: path.join(process.cwd(), '.')
    });

    changelog
        .write({
            githubRepoPath: `https://github.com/${REPO}`,
            lastVersion: previousVersion
        })
        .sort()
        .clean();
};

const previousRelease = () => {
    return releaseUtils
        .releases
        .get({
            userAgent: USER_AGENT,
            uri: `https://api.github.com/repos/${REPO}/releases`
        })
        .then((response) => {
            if (!response || !response.length) {
                console.log('No releases found. Skipping');
                return;
            }
            let prevVersion = response[0].tag_name || response[0].name;
            console.log(`Previous version ${prevVersion}`);
            return prevVersion;
        });
};

/**
 *
 * `yarn ship` will trigger `postship` task.
 *
 * [optional] For full automation
 *
 * `GHOST=2.10.1,2.10.0 yarn ship`
 * First value: Ships with Ghost
 * Second value: Compatible with Ghost/GScan
 *
 * You can manually run in case the task has thrown an error.
 *
 * `npm_package_version=0.5.0 gulp release`
 */
const release = () => {
    // @NOTE: https://yarnpkg.com/lang/en/docs/cli/version/
    // require(./package.json) can run into caching issues, this re-reads from file everytime on release
    var packageJSON = JSON.parse(fs.readFileSync('./package.json'));
    const newVersion = packageJSON.version;
    let shipsWithGhost = '{version}';
    let compatibleWithGhost = '2.10.0';
    const ghostEnvValues = process.env.GHOST || null;

    if (ghostEnvValues) {
        shipsWithGhost = ghostEnvValues.split(',')[0];
        compatibleWithGhost = ghostEnvValues.split(',')[1];

        if (!compatibleWithGhost) {
            compatibleWithGhost = '2.10.0';
        }
    }

    if (!newVersion || newVersion === '') {
        console.log('Invalid version.');
        return;
    }

    console.log(`\nDraft release for ${newVersion}.`);

    if (!config || !config.github || !config.github.username || !config.github.token) {
        console.log('Please copy config.example.json and configure Github token.');
        return;
    }

    return previousRelease()
        .then((previousVersion) => {
            changelog({previousVersion});

            return releaseUtils
                .releases
                .create({
                    draft: true,
                    preRelease: false,
                    tagName: newVersion,
                    releaseName: newVersion,
                    userAgent: USER_AGENT,
                    uri: `https://api.github.com/repos/${REPO}/releases`,
                    github: {
                        username: config.github.username,
                        token: config.github.token
                    },
                    content: [`**Ships with Ghost ${shipsWithGhost} Compatible with Ghost >= ${compatibleWithGhost}**\n\n`],
                    changelogPath: CHANGELOG_PATH
                })
                .then((response) => {
                    console.log(`\nRelease draft generated: ${response.releaseUrl}\n`);
                });
        });
};

exports.release = release;
