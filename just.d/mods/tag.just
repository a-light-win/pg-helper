
[no-cd]
@next-tag mode='patch':
  just -f {{ source_file() }} _tag {{ mode }}

[no-cd]
@next-builder mode='patch':
  just -f {{ source_file() }} _tag {{ mode }} 'builder'

[no-cd,private]
_tag mode='patch' prefix='v':
  #!/usr/bin/env python3
  import sys
  import re
  from subprocess import call, check_output

  prefix = '{{ prefix }}'
  mode = '{{ mode }}'

  version = check_output(["git", "describe", "--tags", "--dirty", "--match", f'{ prefix }*']).decode().strip()

  if 'dirty' in version:
    print("Working tree is dirty, exiting", file=sys.stderr)
    sys.exit(1)

  match = re.match(
    rf'^{ prefix }'
    r"(?P<major>\d+)"
    r"(?:\.(?P<minor>\d+))?"
    r"(?:\.(?P<patch>\d+))?"
    r"(?:-(?P<devstage>alpha|beta|rc)\.(?P<dev>\d+))?"
    r"(?:-(?P<hash>\d+-g[0-9a-f]+))?$",
    version,
  )

  if not match:
    print(f"'{ version }' is not a invalid tag, exiting", file=sys.stderr)
    sys.exit(1)

  if not match.group('hash') and mode not in ['major', 'minor', 'patch', 'devstage']:
    print("There are no commits since the last tag, exiting", file=sys.stderr)
    sys.exit(0)

  devstage = None
  dev = 1
  next_devstage = {
    None: 'alpha',
    'alpha': 'beta',
    'beta': 'rc',
    'rc': 'rc',
  }
  if mode == 'major':
    major = int(match.group('major')) + 1
    minor = 0
    patch = 0
  elif mode == 'minor':
    major = int(match.group('major'))
    minor = int(match.group('minor')) + 1
    patch = 0
  elif mode == 'patch':
    major = int(match.group('major'))
    minor = int(match.group('minor'))
    patch = int(match.group('patch'))
    if not match.group('devstage'):
      patch += 1
  elif mode == 'devstage':
    major = int(match.group('major'))
    minor = int(match.group('minor'))
    patch = int(match.group('patch'))
    devstage = match.group('devstage')
    if devstage is None:
      patch += 1
    elif devstage == 'rc':
      dev = int(match.group('dev')) + 1
    devstage = next_devstage[devstage]
  elif mode == 'dev':
    major = int(match.group('major'))
    minor = int(match.group('minor'))
    patch = int(match.group('patch'))
    devstage = match.group('devstage')
    if devstage is None:
      patch += 1
      devstage = 'alpha'
      dev = 1
    else:
      dev = int(match.group('dev')) + 1
  else:
    print("Invalid mode '{ mode }', exiting", file=sys.stderr)
    sys.exit(1)

  version = f"{ prefix }{ major }.{ minor }.{ patch }"
  if devstage:
    version += f"-{ devstage }.{ dev }"

  print(f'Apply tag: { version }')
  call(["git", "tag", "-s", "-m", f'Version {version} release', version])

