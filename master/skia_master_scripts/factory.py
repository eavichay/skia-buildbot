# Copyright (c) 2012 The Chromium Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

"""Utility class to build the Skia master BuildFactory's.

Based on gclient_factory.py and adds Skia-specific steps."""

import ntpath
import posixpath

from buildbot.process.properties import WithProperties
from master.factory import gclient_factory

from skia_master_scripts import commands as skia_commands
import config

AUTOGEN_SVN_BASEURL = 'https://skia-autogen.googlecode.com/svn'
BENCH_REPEAT_COUNT = 20
BENCH_GRAPH_NUM_REVISIONS = 150
BENCH_GRAPH_X = 1024
BENCH_GRAPH_Y = 768

# TODO(epoger): My intent is to make the build steps identical on all platforms
# and thus remove the need for the whole target_platform parameter.
# For now, these must match the target_platform values used in
# third_party/chromium_buildbot/scripts/master/factory/gclient_factory.py ,
# because we pass these values into GClientFactory.__init__() .
TARGET_PLATFORM_LINUX = 'linux'
TARGET_PLATFORM_MAC = 'mac'
TARGET_PLATFORM_WIN32 = 'win32'

# TODO: This is a workaround for https://code.google.com/p/skia/issues/detail?id=685
# ('gsutil upload fails with "BotoServerError: 500 Internal Server Error", but
# only for certain destination filenames').
# Modify this value to generate graphs with a new filename that will upload
# successfully, when uploading the old filename starts to fail.
GRAPH_PATH_MODIFIER = '-2'

class SkiaFactory(gclient_factory.GClientFactory):
  """Encapsulates data and methods common to the Skia master.cfg files."""

  def __init__(self, do_upload_results=False,
               build_subdir='trunk', other_subdirs=None,
               target_platform=None, configuration='Debug',
               default_timeout=600,
               environment_variables=None, gm_image_subdir=None,
               perf_output_basedir=None, builder_name=None):
    """Instantiates a SkiaFactory as appropriate for this target_platform.

    do_upload_results: whether we should upload bench/gm results
    build_subdir: subdirectory to check out and then build within
    other_subdirs: list of other subdirectories to also check out (or None)
    target_platform: a string such as TARGET_PLATFORM_LINUX
    configuration: 'Debug' or 'Release'
    default_timeout: default timeout for each command, in seconds
    environment_variables: dictionary of environment variables that should
        be passed to all commands
    gm_image_subdir: directory containing images for comparison against results
        of gm tool
    perf_output_basedir: path to directory under which to store performance
        data, or None if we don't want to store performance data
    builder_name: name of the builder associated with this factory
    """
    # Create gclient solutions corresponding to the main build_subdir
    # and other directories we also wish to check out.
    solutions = [gclient_factory.GClientSolution(
        svn_url=config.Master.skia_url + build_subdir, name=build_subdir)]
    if other_subdirs:
      for other_subdir in other_subdirs:
        solutions.append(gclient_factory.GClientSolution(
            svn_url=config.Master.skia_url + other_subdir, name=other_subdir))
    gclient_factory.GClientFactory.__init__(
        self, build_dir='', solutions=solutions,
        target_platform=target_platform)

    self._do_upload_results = do_upload_results
    self._configuration = configuration
    self._factory = self.BaseFactory(factory_properties=None)
    self._gm_image_subdir = gm_image_subdir
    self._builder_name = builder_name
    self._target_platform = target_platform

    # Set _default_clobber based on config.Master
    self._default_clobber = getattr(config.Master, 'default_clobber', False)

    # Platform-specific stuff.
    if target_platform == TARGET_PLATFORM_WIN32:
      self.TargetPathJoin = ntpath.join
      self._make_flags = 'BUILDTYPE=%s' % self._configuration
    else:
      self.TargetPathJoin = posixpath.join
      self._make_flags = '--jobs --max-load=4.0 BUILDTYPE=%s' % (
          self._configuration)

    # Figure out where we are going to store performance output.
    if perf_output_basedir:
      self._perf_data_dir = self.TargetPathJoin(
          perf_output_basedir, builder_name, 'data')
      self._perf_graphs_dir = self.TargetPathJoin(
          perf_output_basedir, builder_name, 'graphs')
    else:
      self._perf_data_dir = None
      self._perf_graphs_dir = None

    # Figure out where we are going to store images generated by GM.
    self._gm_actual_basedir = self.TargetPathJoin('..', '..', 'gm', 'actual')
    self._gm_merge_basedir = self.TargetPathJoin('..', '..', 'gm', 'merge')
    self._gm_actual_svn_baseurl = '%s/%s' % (AUTOGEN_SVN_BASEURL, 'gm-actual')
    self._autogen_svn_username_file = '.autogen_svn_username'
    self._autogen_svn_password_file = '.autogen_svn_password'

    # Get an implementation of SkiaCommands as appropriate for
    # this target_platform.
    workdir = self.TargetPathJoin('build', build_subdir)
    self._skia_cmd_obj = skia_commands.SkiaCommands(
        target_platform=target_platform, factory=self._factory,
        configuration=configuration, workdir=workdir,
        target_arch=None, default_timeout=default_timeout,
        environment_variables=environment_variables)

  def _Make(self, target, description):
    """Build a single target."""
    self._skia_cmd_obj.AddRunCommand(command='make %s %s' % (target, self._make_flags),
                                     description=description)

  def Compile(self):
    """Compile step.  Build everything. """
    self._Make('core',  'BuildCore')
    self._Make('tests', 'BuildTests')
    self._Make('gm',    'BuildGM')
    self._Make('tools', 'BuildTools')
    self._Make('bench', 'BuildBench')
    self._Make('all',   'BuildAllOtherTargets')

  def RunTests(self):
    """ Run the unit tests. """
    self._skia_cmd_obj.AddRunCommand(
        command=self.TargetPathJoin('out', self._configuration, 'tests'),
        description='RunTests')

  def PrepForGM(self, path_to_gm, gm_actual_dir):
    """ Ensure that the output image directory for GM exists and is empty. """
    if self._target_platform == TARGET_PLATFORM_WIN32:
      # This ridiculous hack is the simplest way I could think of to
      # consistently create an empty directory (whether it already existed
      # or not) on the Windows command line.
      command_list = [
          'mkdir %s' % self.TargetPathJoin(gm_actual_dir, 'bogus-subdir'),
          'rmdir /s /q %s' % gm_actual_dir,
          'mkdir %s' % gm_actual_dir,
          ]
    else:
      command_list = [
          'rm -rf %s' % gm_actual_dir,
          'mkdir -p %s' % gm_actual_dir,
          ]
    self._skia_cmd_obj.AddRunCommandList(
        command_list=command_list, description='PrepForGM')

  def RunGM(self, path_to_gm, gm_actual_dir):
    """ Run the "GM" tool, saving the images to disk. """
    gm_command = '%s -w %s' % (path_to_gm, gm_actual_dir)
    self._skia_cmd_obj.AddRunCommand(
        command=gm_command,
        description='GenerateGMs')

  def CompareGMs(self, gm_actual_dir):
    """ Run the "skdiff" tool to compare the "actual" GM images we just
    generated to the baselines in _gm_image_subdir. """
    path_to_skdiff = self.TargetPathJoin('out', self._configuration, 'skdiff')
    gm_expected_dir = self.TargetPathJoin('gm', self._gm_image_subdir)
    self._skia_cmd_obj.AddRunCommand(
        command='%s --listfilenames --nodiffs --nomatch README'
                ' --failonresult DifferentPixels --failonresult DifferentSizes'
                ' --failonresult DifferentOther  --failonresult Unknown'
                ' %s %s' % (path_to_skdiff, gm_expected_dir, gm_actual_dir),
        description='CompareGMs')

  def RunBench(self):
    """ Run "bench", piping the output somewhere so we can graph
    results over time. """

    # TODO(epoger): Currently this is a hack--we just tell the slave to
    # pipe the output to a directory on local disk.
    # Eventually, we will want the master to capture the output and store it.
    #
    # Running bench can be quite slow, so run it fewer times if we aren't
    # recording the output.
    if self._perf_data_dir:
      count = BENCH_REPEAT_COUNT
    else:
      count = 1
    path_to_bench = self.TargetPathJoin('out', self._configuration, 'bench')
    base_command = '%s -repeat %d -timers wcg -logPerIter 1' % (
        path_to_bench, count)
    if self._perf_data_dir:
      # TODO(epoger): For now, this relies on AddRunCommandList() wrapping all
      # command in WithProperties().
      data_file = self.TargetPathJoin(
          self._perf_data_dir, 'bench_r%%(%s:-)s_data' % 'revision')

      if self._target_platform == TARGET_PLATFORM_WIN32:
        command_list = [
            '(if not exist %s (mkdir %s))' % (
                self._perf_data_dir, self._perf_data_dir),
            '%s > %s' % (base_command, data_file),
            ]
      else:
        command_list = [
            'mkdir -p %s' % self._perf_data_dir,
            '%s | tee %s' % (base_command, data_file),
            'chmod a+r %s' % data_file,
            ]
    else:
      command_list = [base_command]
    self._skia_cmd_obj.AddRunCommandList(
        command_list=command_list, description='RunBench', timeout=1200)

  def BenchGraphs(self):
    """ Generate and upload bench performance graphs (but only if we have been
    recording bench output for this build type). """
    if self._perf_data_dir:
      path_to_bench_graph_svg = self.TargetPathJoin(
          'bench', 'bench_graph_svg.py')
      graph_title = 'Bench_Performance_for_%s' % self._builder_name
      graph_filepath = self.TargetPathJoin(
          self._perf_graphs_dir, 'graph-%s%s.xhtml' % (
              self._builder_name, GRAPH_PATH_MODIFIER))
      if self._target_platform == TARGET_PLATFORM_WIN32:
        mkdir_command = '(if not exist %s (mkdir %s))' % (
            self._perf_graphs_dir, self._perf_graphs_dir)
      else:
        mkdir_command = 'mkdir -p %s' % self._perf_graphs_dir
      gen_command = 'python %s -d %s -r -%d -f -%d -x %d -y %d -l %s -o %s' % (
          path_to_bench_graph_svg, self._perf_data_dir,
          BENCH_GRAPH_NUM_REVISIONS, BENCH_GRAPH_NUM_REVISIONS,
          BENCH_GRAPH_X, BENCH_GRAPH_Y, graph_title, graph_filepath)
      command_list = [mkdir_command, gen_command]
      self._skia_cmd_obj.AddRunCommandList(
          command_list=command_list, description='GenerateBenchGraphs')
      if self._do_upload_results:
        self._skia_cmd_obj.AddUploadToBucket(
            source_filepath=graph_filepath, description='UploadBenchGraphs')

  def Build(self, clobber=None):
    """Build and return the complete BuildFactory.

    clobber: boolean indicating whether we should clean before building
    """
    # Do all the build steps first, so we will find out about build breakages
    # as soon as possible.
    if clobber is None:
      clobber = self._default_clobber
    if clobber:
      self._skia_cmd_obj.AddClean()
    self.Compile()
    self.RunTests()
    path_to_gm = self.TargetPathJoin('out', self._configuration, 'gm')
    gm_actual_dir = self.TargetPathJoin(
        self._gm_actual_basedir, self._gm_image_subdir)
    self.PrepForGM(path_to_gm, gm_actual_dir)
    self.RunGM(path_to_gm, gm_actual_dir)
    if self._do_upload_results:
      self._skia_cmd_obj.AddUploadGMResults(
          gm_image_subdir=self._gm_image_subdir,
          builder_name=self._builder_name)
    self.CompareGMs(gm_actual_dir)
    self.RunBench()
    self.BenchGraphs()

    return self._factory
