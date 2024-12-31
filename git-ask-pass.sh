#!/bin/sh
#
# Copyright (c) 2024. Devtron Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

# This script is used as the command supplied to GIT_ASKPASS as a way to supply username/password
# credentials to git, without having to use git credentials helpers, or having on-disk config.
case "$1" in
Username*) echo "${GIT_USERNAME}" ;;
Password*) echo "${GIT_PASSWORD}" ;;
esac
