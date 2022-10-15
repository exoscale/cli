const completionSpec: Fig.Spec = {
  name: 'exo',
  description: 'Manage your Exoscale infrastructure easily',
  subcommands: [
    {
      name: ['completion'],
      description:
        'Generate the autocompletion script for the specified shell',
      subcommands: [
        {
          name: ['bash'],
          description: 'Generate the autocompletion script for bash',
          options: [
            {
              name: ['--no-descriptions'],
              description: 'disable completion descriptions',
            },
          ],
        },
        {
          name: ['fish'],
          description: 'Generate the autocompletion script for fish',
          options: [
            {
              name: ['--no-descriptions'],
              description: 'disable completion descriptions',
            },
          ],
        },
        {
          name: ['powershell'],
          description:
            'Generate the autocompletion script for powershell',
          options: [
            {
              name: ['--no-descriptions'],
              description: 'disable completion descriptions',
            },
          ],
        },
        {
          name: ['zsh'],
          description: 'Generate the autocompletion script for zsh',
          options: [
            {
              name: ['--no-descriptions'],
              description: 'disable completion descriptions',
            },
          ],
        },
      ],
    },
    {
      name: ['c', 'compute'],
      description: 'Compute services management',
      subcommands: [
        {
          name: ['aag', 'anti-affinity-group'],
          description: 'Anti-Affinity Groups management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create an Anti-Affinity Group',
              options: [
                {
                  name: ['--description'],
                  description: 'Anti-Affinity Group description',
                  args: [{ name: 'description' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete an Anti-Affinity Group',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Anti-Affinity Groups',
            },
            {
              name: ['get', 'show'],
              description: 'Show an Anti-Affinity Group details',
            },
          ],
        },
        {
          name: ['dt', 'deploy-target'],
          description: 'Compute instance Deploy Targets management',
          subcommands: [
            {
              name: ['list'],
              description: 'List Deploy Targets',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Deploy Target details',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'Deploy Target zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['eip', 'elastic-ip'],
          description: 'Elastic IP addresses management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create an Elastic IP',
              options: [
                {
                  name: ['--description'],
                  description: 'Elastic IP description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--healthcheck-interval'],
                  description:
                    'managed Elastic IP health checking interval in seconds',
                  args: [
                    { name: 'healthcheck-interval', default: '10' },
                  ],
                },
                {
                  name: ['--healthcheck-mode'],
                  description:
                    'managed Elastic IP health checking mode (tcp|http|https)',
                  args: [{ name: 'healthcheck-mode' }],
                },
                {
                  name: ['--healthcheck-port'],
                  description:
                    'managed Elastic IP health checking port',
                  args: [{ name: 'healthcheck-port', default: '0' }],
                },
                {
                  name: ['--healthcheck-strikes-fail'],
                  description:
                    'number of failed attempts before considering a managed Elastic IP health check unhealthy',
                  args: [
                    { name: 'healthcheck-strikes-fail', default: '2' },
                  ],
                },
                {
                  name: ['--healthcheck-strikes-ok'],
                  description:
                    'number of successful attempts before considering a managed Elastic IP health check healthy',
                  args: [
                    { name: 'healthcheck-strikes-ok', default: '3' },
                  ],
                },
                {
                  name: ['--healthcheck-timeout'],
                  description:
                    'managed Elastic IP health checking timeout in seconds',
                  args: [{ name: 'healthcheck-timeout', default: '3' }],
                },
                {
                  name: ['--healthcheck-tls-skip-verify'],
                  description:
                    'disable TLS certificate verification for managed Elastic IP health checking in https mode',
                },
                {
                  name: ['--healthcheck-tls-sni'],
                  description:
                    'managed Elastic IP health checking server name to present with SNI in https mode',
                  args: [{ name: 'healthcheck-tls-sni' }],
                },
                {
                  name: ['--healthcheck-uri'],
                  description:
                    'managed Elastic IP health checking URI (required in http(s) mode)',
                  args: [{ name: 'healthcheck-uri' }],
                },
                {
                  name: ['--ipv6'],
                  description: 'create Elastic IPv6 prefix',
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Elastic IP zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete an Elastic IP',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Elastic IP zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Elastic IPs',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show an Elastic IP details',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'Elastic IP zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update an Elastic IP',
              options: [
                {
                  name: ['--description'],
                  description: 'Elastic IP description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--healthcheck-interval'],
                  description:
                    'managed Elastic IP health checking interval in seconds',
                  args: [
                    { name: 'healthcheck-interval', default: '0' },
                  ],
                },
                {
                  name: ['--healthcheck-mode'],
                  description:
                    'managed Elastic IP health checking mode (tcp|http|https)',
                  args: [{ name: 'healthcheck-mode' }],
                },
                {
                  name: ['--healthcheck-port'],
                  description:
                    'managed Elastic IP health checking port',
                  args: [{ name: 'healthcheck-port', default: '0' }],
                },
                {
                  name: ['--healthcheck-strikes-fail'],
                  description:
                    'number of failed attempts before considering a managed Elastic IP health check unhealthy',
                  args: [
                    { name: 'healthcheck-strikes-fail', default: '0' },
                  ],
                },
                {
                  name: ['--healthcheck-strikes-ok'],
                  description:
                    'number of successful attempts before considering a managed Elastic IP health check healthy',
                  args: [
                    { name: 'healthcheck-strikes-ok', default: '0' },
                  ],
                },
                {
                  name: ['--healthcheck-timeout'],
                  description:
                    'managed Elastic IP health checking timeout in seconds',
                  args: [{ name: 'healthcheck-timeout', default: '0' }],
                },
                {
                  name: ['--healthcheck-tls-skip-verify'],
                  description:
                    'disable TLS certificate verification for managed Elastic IP health checking in https mode',
                },
                {
                  name: ['--healthcheck-tls-sni'],
                  description:
                    'managed Elastic IP health checking server name to present with SNI in https mode',
                  args: [{ name: 'healthcheck-tls-sni' }],
                },
                {
                  name: ['--healthcheck-uri'],
                  description:
                    'managed Elastic IP health checking URI (required in http(s) mode)',
                  args: [{ name: 'healthcheck-uri' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Elastic IP zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['i', 'instance'],
          description: 'Compute instances management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create a Compute instance',
              options: [
                {
                  name: ['--anti-affinity-group'],
                  description:
                    'instance Anti-Affinity Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'anti-affinity-group' }],
                },
                {
                  name: ['--cloud-init'],
                  description:
                    'instance cloud-init user data configuration file path',
                  args: [{ name: 'cloud-init' }],
                },
                {
                  name: ['--cloud-init-compress'],
                  description: 'compress instance cloud-init user data',
                },
                {
                  name: ['--deploy-target'],
                  description: 'instance Deploy Target NAME|ID',
                  args: [{ name: 'deploy-target' }],
                },
                {
                  name: ['--disk-size'],
                  description: 'instance disk size',
                  args: [{ name: 'disk-size', default: '50' }],
                },
                {
                  name: ['--instance-type'],
                  description: 'instance type (format: [FAMILY.]SIZE)',
                  args: [
                    {
                      name: 'instance-type',
                      default: 'standard.medium',
                    },
                  ],
                },
                {
                  name: ['--ipv6'],
                  description: 'enable IPv6 on instance',
                },
                {
                  name: ['--label'],
                  description: 'instance label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--private-network'],
                  description:
                    'instance Private Network NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'private-network' }],
                },
                {
                  name: ['--security-group'],
                  description:
                    'instance Security Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'security-group' }],
                },
                {
                  name: ['--ssh-key'],
                  description: 'SSH key to deploy on the instance',
                  args: [{ name: 'ssh-key' }],
                },
                {
                  name: ['--template'],
                  description: 'instance template NAME|ID',
                  args: [{ name: 'template' }],
                },
                {
                  name: ['--template-visibility'],
                  description:
                    'instance template visibility (public|private)',
                  args: [
                    { name: 'template-visibility', default: 'public' },
                  ],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete a Compute instance',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['eip', 'elastic-ip'],
              description:
                'Manage Compute instance Elastic IP addresses',
              subcommands: [
                {
                  name: ['attach'],
                  description:
                    'Attach an Elastic IP to a Compute instance',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['detach'],
                  description:
                    'Detach a Compute instance from a Elastic IP',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Compute instances',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['privnet', 'private-network'],
              description: 'Manage Compute instance Private Networks',
              subcommands: [
                {
                  name: ['attach'],
                  description:
                    'Attach a Compute instance to a Private Network',
                  options: [
                    {
                      name: ['--ip'],
                      description:
                        'network IP address to assign to the Compute instance (managed Private Networks only)',
                      args: [{ name: 'ip' }],
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['detach'],
                  description:
                    'Detach a Compute instance from a Private Network',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['update-ip'],
                  description:
                    'Update a Compute instance Private Network IP address',
                  options: [
                    {
                      name: ['--ip'],
                      description:
                        'network IP address to assign to the Compute instance',
                      args: [{ name: 'ip' }],
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
              ],
            },
            {
              name: ['reboot'],
              description: 'Reboot a Compute instance',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['reset'],
              description: 'Reset a Compute instance',
              options: [
                {
                  name: ['--disk-size'],
                  description:
                    'disk size to reset the instance to (default: current instance disk size)',
                  args: [{ name: 'disk-size', default: '0' }],
                },
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--template'],
                  description:
                    'template NAME|ID to reset the instance to (default: current instance template)',
                  args: [{ name: 'template' }],
                },
                {
                  name: ['--template-visibility'],
                  description:
                    'instance template visibility (public|private)',
                  args: [
                    { name: 'template-visibility', default: 'public' },
                  ],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['resize-disk'],
              description: 'Resize a Compute instance disk',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['scale'],
              description: 'Scale a Compute instance',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['scp'],
              description: 'SCP files to/from a Compute instance',
              options: [
                { name: ['--ipv6', '-6'], description: '' },
                {
                  name: ['--login', '-l'],
                  description: '',
                  args: [{ name: 'login' }],
                },
                {
                  name: ['--print-command'],
                  description:
                    'print the SCP command that would be executed instead of executing it',
                },
                {
                  name: ['--recursive', '-r'],
                  description: 'recursively copy entire directories',
                },
                {
                  name: ['--replace-str', '-i'],
                  description:
                    'string to replace with the actual Compute instance information (i.e. username@IP-ADDRESS)',
                  args: [{ name: 'replace-str', default: '{}' }],
                },
                {
                  name: ['--scp-options', '-o'],
                  description:
                    'additional options to pass to the scp(1) command',
                  args: [{ name: 'scp-options' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['sg', 'security-group'],
              description: 'Manage Compute instance Security Groups',
              subcommands: [
                {
                  name: ['add'],
                  description:
                    'Add a Compute instance to Security Groups',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['rm', 'remove'],
                  description:
                    'Remove a Compute instance from Security Groups',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Compute instance details',
              options: [
                {
                  name: ['--user-data', '-u'],
                  description:
                    'show instance cloud-init user data configuration',
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['snap', 'snapshot'],
              description: 'Manage Compute instance snapshots',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create a Compute instance snapshot',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'instance zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Compute instance snapshot',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'snapshot zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['export'],
                  description: 'Export a Compute instance snapshot',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'snapshot zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['list'],
                  description: 'List Compute instance snapshots',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'zone to filter results to',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['revert'],
                  description:
                    'Revert a Compute instance to a snapshot',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'snapshot zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['get', 'show'],
                  description:
                    'Show a Compute instance snapshot details',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'snapshot zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
              ],
            },
            {
              name: ['ssh'],
              description: 'Log into a Compute instance via SSH',
              options: [
                { name: ['--ipv6', '-6'], description: '' },
                {
                  name: ['--login', '-l'],
                  description: '',
                  args: [{ name: 'login' }],
                },
                {
                  name: ['--print-command'],
                  description:
                    'print the SSH command that would be executed instead of executing it',
                },
                {
                  name: ['--print-ssh-config'],
                  description:
                    'print the corresponding SSH information in a format compatible with ssh_config(5)',
                },
                {
                  name: ['--ssh-options', '-o'],
                  description:
                    'additional options to pass to the ssh(1) command',
                  args: [{ name: 'ssh-options' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['start'],
              description: 'Start a Compute instance',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--rescue-profile'],
                  description:
                    'rescue profile to start the instance with',
                  args: [{ name: 'rescue-profile' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['stop'],
              description: 'Stop a Compute instance',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update an Instance ',
              options: [
                {
                  name: ['--cloud-init', '-c'],
                  description:
                    'instance cloud-init user data configuration file path',
                  args: [{ name: 'cloud-init' }],
                },
                {
                  name: ['--cloud-init-compress'],
                  description: 'compress instance cloud-init user data',
                },
                {
                  name: ['--label'],
                  description: 'instance label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'instance name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'instance zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['pool', 'instance-pool'],
          description: 'Instance Pools management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create an Instance Pool',
              options: [
                {
                  name: ['--anti-affinity-group', '-a'],
                  description:
                    'managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'anti-affinity-group' }],
                },
                {
                  name: ['--cloud-init', '-c'],
                  description:
                    'cloud-init user data configuration file path',
                  args: [{ name: 'cloud-init' }],
                },
                {
                  name: ['--cloud-init-compress'],
                  description: 'compress instance cloud-init user data',
                },
                {
                  name: ['--deploy-target'],
                  description:
                    'managed Compute instances Deploy Target NAME|ID',
                  args: [{ name: 'deploy-target' }],
                },
                {
                  name: ['--description'],
                  description: 'Instance Pool description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--disk', '-d'],
                  description: '[DEPRECATED] use --disk-size',
                  args: [{ name: 'disk', default: '0' }],
                },
                {
                  name: ['--disk-size'],
                  description: 'managed Compute instances disk size',
                  args: [{ name: 'disk-size', default: '50' }],
                },
                {
                  name: ['--elastic-ip', '-e'],
                  description:
                    'managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'elastic-ip' }],
                },
                {
                  name: ['--instance-prefix'],
                  description:
                    'string to prefix managed Compute instances names with',
                  args: [{ name: 'instance-prefix' }],
                },
                {
                  name: ['--instance-type'],
                  description:
                    'managed Compute instances type (format: [FAMILY.]SIZE)',
                  args: [
                    {
                      name: 'instance-type',
                      default: 'standard.medium',
                    },
                  ],
                },
                {
                  name: ['--ipv6', '-6'],
                  description:
                    'enable IPv6 on managed Compute instances',
                },
                {
                  name: ['--keypair', '-k'],
                  description: '[DEPRECATED] use --ssh-key',
                  args: [{ name: 'keypair' }],
                },
                {
                  name: ['--label'],
                  description:
                    'Instance Pool label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--private-network'],
                  description:
                    'managed Compute instances Private Network NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'private-network' }],
                },
                {
                  name: ['--privnet', '-p'],
                  description: '[DEPRECATED] use --private-network',
                  isRepeatable: true,
                  args: [{ name: 'privnet' }],
                },
                {
                  name: ['--security-group', '-s'],
                  description:
                    'managed Compute instances Security Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'security-group' }],
                },
                {
                  name: ['--service-offering', '-o'],
                  description: '[DEPRECATED] use --instance-type',
                  args: [{ name: 'service-offering' }],
                },
                {
                  name: ['--size'],
                  description: 'Instance Pool size',
                  args: [{ name: 'size', default: '1' }],
                },
                {
                  name: ['--ssh-key'],
                  description:
                    'SSH key to deploy on managed Compute instances',
                  args: [{ name: 'ssh-key' }],
                },
                {
                  name: ['--template', '-t'],
                  description:
                    'managed Compute instances template NAME|ID',
                  args: [{ name: 'template' }],
                },
                {
                  name: ['--template-filter'],
                  description:
                    'managed Compute instances template filter',
                  args: [
                    { name: 'template-filter', default: 'featured' },
                  ],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Instance Pool zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete an Instance Pool',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Instance Pool zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['evict'],
              description: 'Evict Instance Pool members',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Instance Pool zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Instance Pools',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['scale'],
              description: 'Scale an Instance Pool size',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Instance Pool zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show an Instance Pool details',
              options: [
                {
                  name: ['--user-data', '-u'],
                  description:
                    'show cloud-init user data configuration',
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Instance Pool zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update an Instance Pool',
              options: [
                {
                  name: ['--anti-affinity-group', '-a'],
                  description:
                    'managed Compute instances Anti-Affinity Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'anti-affinity-group' }],
                },
                {
                  name: ['--cloud-init', '-c'],
                  description:
                    'cloud-init user data configuration file path',
                  args: [{ name: 'cloud-init' }],
                },
                {
                  name: ['--cloud-init-compress'],
                  description: 'compress instance cloud-init user data',
                },
                {
                  name: ['--deploy-target'],
                  description:
                    'managed Compute instances Deploy Target NAME|ID',
                  args: [{ name: 'deploy-target' }],
                },
                {
                  name: ['--description'],
                  description: 'Instance Pool description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--disk', '-d'],
                  description: '[DEPRECATED] use --disk-size',
                  args: [{ name: 'disk', default: '0' }],
                },
                {
                  name: ['--disk-size'],
                  description: 'managed Compute instances disk size',
                  args: [{ name: 'disk-size', default: '0' }],
                },
                {
                  name: ['--elastic-ip', '-e'],
                  description:
                    'managed Compute instances Elastic IP ADDRESS|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'elastic-ip' }],
                },
                {
                  name: ['--instance-prefix'],
                  description:
                    'string to prefix managed Compute instances names with',
                  args: [{ name: 'instance-prefix' }],
                },
                {
                  name: ['--instance-type'],
                  description:
                    'managed Compute instances type (format: [FAMILY.]SIZE)',
                  args: [{ name: 'instance-type' }],
                },
                {
                  name: ['--ipv6', '-6'],
                  description:
                    'enable IPv6 on managed Compute instances',
                },
                {
                  name: ['--keypair', '-k'],
                  description: '[DEPRECATED] use --ssh-key',
                  args: [{ name: 'keypair' }],
                },
                {
                  name: ['--label'],
                  description:
                    'Instance Pool label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Instance Pool name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--private-network'],
                  description:
                    'managed Compute instances Private Network NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'private-network' }],
                },
                {
                  name: ['--privnet', '-p'],
                  description: '[DEPRECATED] use --private-network',
                  isRepeatable: true,
                  args: [{ name: 'privnet' }],
                },
                {
                  name: ['--security-group', '-s'],
                  description:
                    'managed Compute instances Security Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'security-group' }],
                },
                {
                  name: ['--service-offering', '-o'],
                  description: '[DEPRECATED] use --instance-type',
                  args: [{ name: 'service-offering' }],
                },
                {
                  name: ['--size'],
                  description:
                    "[DEPRECATED] use the 'exo compute instance-pool scale' command",
                  args: [{ name: 'size', default: '0' }],
                },
                {
                  name: ['--ssh-key'],
                  description:
                    'SSH key to deploy on managed Compute instances',
                  args: [{ name: 'ssh-key' }],
                },
                {
                  name: ['--template', '-t'],
                  description:
                    'managed Compute instances template NAME|ID',
                  args: [{ name: 'template' }],
                },
                {
                  name: ['--template-filter'],
                  description:
                    'managed Compute instances template filter',
                  args: [
                    { name: 'template-filter', default: 'featured' },
                  ],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Instance Pool zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['template', 'instance-template'],
          description: 'Compute instance templates management',
          subcommands: [
            {
              name: ['rm', 'delete'],
              description: 'Delete a Compute instance template',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'template zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['list'],
              description: 'List Compute instance templates',
              options: [
                {
                  name: ['--family', '-f'],
                  description: 'template family to filter results to',
                  args: [{ name: 'family' }],
                },
                {
                  name: ['--visibility', '-v'],
                  description: 'template visibility (public|private)',
                  args: [{ name: 'visibility', default: 'public' }],
                },
                {
                  name: ['--zone', '-z'],
                  description:
                    "zone to filter results to (default: current account's default zone)",
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['add', 'register'],
              description: 'Register a new Compute instance template',
              options: [
                {
                  name: ['--boot-mode'],
                  description: 'template boot mode (legacy|uefi)',
                  args: [{ name: 'boot-mode', default: 'legacy' }],
                },
                {
                  name: ['--build'],
                  description: 'template build',
                  args: [{ name: 'build' }],
                },
                {
                  name: ['--description'],
                  description: 'template description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--disable-password'],
                  description: 'disable password-based authentication',
                },
                {
                  name: ['--disable-ssh-key'],
                  description: 'disable SSH key-based authentication',
                },
                {
                  name: ['--from-snapshot'],
                  description:
                    'ID of a Compute instance snapshot to register as template',
                  args: [{ name: 'from-snapshot' }],
                },
                {
                  name: ['--maintainer'],
                  description: 'template maintainer',
                  args: [{ name: 'maintainer' }],
                },
                {
                  name: ['--timeout'],
                  description:
                    'registration timeout duration in seconds',
                  args: [{ name: 'timeout', default: '3600' }],
                },
                {
                  name: ['--username'],
                  description: 'template default username',
                  args: [{ name: 'username' }],
                },
                {
                  name: ['--version'],
                  description: 'template version',
                  args: [{ name: 'version' }],
                },
                {
                  name: ['--zone', '-z'],
                  description:
                    "zone to register the template into (default: current account's default zone)",
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Compute instance template details',
              options: [
                {
                  name: ['--visibility', '-v'],
                  description: 'template visibility (public|private)',
                  args: [{ name: 'visibility', default: 'public' }],
                },
                {
                  name: ['--zone', '-z'],
                  description:
                    "zone to filter results to (default: current account's default zone)",
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['instance-type'],
          description: 'Compute instance types management',
          subcommands: [
            {
              name: ['list'],
              description: 'List Compute instance types',
              options: [{ name: ['--verbose', '-v'], description: '' }],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Compute instance type details',
            },
          ],
        },
        {
          name: ['nlb', 'load-balancer'],
          description: 'Network Load Balancers management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create a Network Load Balancer',
              options: [
                {
                  name: ['--description'],
                  description: 'Network Load Balancer description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--label'],
                  description:
                    'Network Load Balancer label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Network Load Balancer zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete a Network Load Balancer',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Network Load Balancer zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Network Load Balancers',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['svc', 'service'],
              description: 'Manage Network Load Balancer services',
              subcommands: [
                {
                  name: ['add'],
                  description:
                    'Add a service to a Network Load Balancer',
                  options: [
                    {
                      name: ['--description'],
                      description: 'service description',
                      args: [{ name: 'description' }],
                    },
                    {
                      name: ['--healthcheck-interval'],
                      description:
                        'service health checking interval in seconds',
                      args: [
                        { name: 'healthcheck-interval', default: '10' },
                      ],
                    },
                    {
                      name: ['--healthcheck-mode'],
                      description:
                        'service health checking mode (tcp|http|https)',
                      args: [
                        { name: 'healthcheck-mode', default: 'tcp' },
                      ],
                    },
                    {
                      name: ['--healthcheck-port'],
                      description:
                        'service health checking port (defaults to target port)',
                      args: [
                        { name: 'healthcheck-port', default: '0' },
                      ],
                    },
                    {
                      name: ['--healthcheck-retries'],
                      description: 'service health checking retries',
                      args: [
                        { name: 'healthcheck-retries', default: '1' },
                      ],
                    },
                    {
                      name: ['--healthcheck-timeout'],
                      description:
                        'service health checking timeout in seconds',
                      args: [
                        { name: 'healthcheck-timeout', default: '5' },
                      ],
                    },
                    {
                      name: ['--healthcheck-tls-sni'],
                      description:
                        'service health checking server name to present with SNI in https mode',
                      args: [{ name: 'healthcheck-tls-sni' }],
                    },
                    {
                      name: ['--healthcheck-uri'],
                      description:
                        'service health checking URI (required in http(s) mode)',
                      args: [{ name: 'healthcheck-uri' }],
                    },
                    {
                      name: ['--instance-pool'],
                      description:
                        'name or ID of the Instance Pool to forward traffic to',
                      args: [{ name: 'instance-pool' }],
                    },
                    {
                      name: ['--port'],
                      description: 'service port',
                      args: [{ name: 'port', default: '0' }],
                    },
                    {
                      name: ['--protocol'],
                      description: 'service network protocol (tcp|udp)',
                      args: [{ name: 'protocol', default: 'tcp' }],
                    },
                    {
                      name: ['--strategy'],
                      description:
                        'load balancing strategy (round-robin|source-hash)',
                      args: [
                        { name: 'strategy', default: 'round-robin' },
                      ],
                    },
                    {
                      name: ['--target-port'],
                      description:
                        'port to forward traffic to on target instances (defaults to service port)',
                      args: [{ name: 'target-port', default: '0' }],
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'Network Load Balancer zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Network Load Balancer service',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'Network Load Balancer zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['get', 'show'],
                  description:
                    'Show a Network Load Balancer service details',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'Network Load Balancer zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['update'],
                  description: 'Update a Network Load Balancer service',
                  options: [
                    {
                      name: ['--description'],
                      description: 'service description',
                      args: [{ name: 'description' }],
                    },
                    {
                      name: ['--healthcheck-interval'],
                      description:
                        'service health checking interval in seconds',
                      args: [
                        { name: 'healthcheck-interval', default: '0' },
                      ],
                    },
                    {
                      name: ['--healthcheck-mode'],
                      description:
                        'service health checking mode (tcp|http|https)',
                      args: [{ name: 'healthcheck-mode' }],
                    },
                    {
                      name: ['--healthcheck-port'],
                      description: 'service health checking port',
                      args: [
                        { name: 'healthcheck-port', default: '0' },
                      ],
                    },
                    {
                      name: ['--healthcheck-retries'],
                      description: 'service health checking retries',
                      args: [
                        { name: 'healthcheck-retries', default: '0' },
                      ],
                    },
                    {
                      name: ['--healthcheck-timeout'],
                      description:
                        'service health checking timeout in seconds',
                      args: [
                        { name: 'healthcheck-timeout', default: '0' },
                      ],
                    },
                    {
                      name: ['--healthcheck-tls-sni'],
                      description:
                        'service health checking server name to present with SNI in https mode',
                      args: [{ name: 'healthcheck-tls-sni' }],
                    },
                    {
                      name: ['--healthcheck-uri'],
                      description:
                        'service health checking URI (required in http(s) mode)',
                      args: [{ name: 'healthcheck-uri' }],
                    },
                    {
                      name: ['--name'],
                      description: 'service name',
                      args: [{ name: 'name' }],
                    },
                    {
                      name: ['--port'],
                      description: 'service port',
                      args: [{ name: 'port', default: '0' }],
                    },
                    {
                      name: ['--protocol'],
                      description: 'service network protocol (tcp|udp)',
                      args: [{ name: 'protocol' }],
                    },
                    {
                      name: ['--strategy'],
                      description:
                        'load balancing strategy (round-robin|source-hash)',
                      args: [{ name: 'strategy' }],
                    },
                    {
                      name: ['--target-port'],
                      description:
                        'port to forward traffic to on target instances',
                      args: [{ name: 'target-port', default: '0' }],
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'Network Load Balancer zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Network Load Balancer details',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'Network Load Balancer zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update a Network Load Balancer',
              options: [
                {
                  name: ['--description'],
                  description: 'Network Load Balancer description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--label'],
                  description:
                    'Network Load Balancer label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--name'],
                  description: 'Network Load Balancer name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Network Load Balancer zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['privnet', 'private-network'],
          description: 'Private Networks management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create a Private Network',
              options: [
                {
                  name: ['--description'],
                  description: 'Private Network description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--end-ip'],
                  description:
                    'managed Private Network range end IP address',
                  args: [{ name: 'end-ip' }],
                },
                {
                  name: ['--netmask'],
                  description: 'managed Private Network netmask',
                  args: [{ name: 'netmask' }],
                },
                {
                  name: ['--start-ip'],
                  description:
                    'managed Private Network range start IP address',
                  args: [{ name: 'start-ip' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Private Network zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete a Private Network',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Private Network zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Private Networks',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Private Network details',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'Private Network zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update a Private Network',
              options: [
                {
                  name: ['--description'],
                  description: 'Private Network description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--end-ip'],
                  description:
                    'managed Private Network range end IP address',
                  args: [{ name: 'end-ip' }],
                },
                {
                  name: ['--name'],
                  description: 'Private Network name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--netmask'],
                  description: 'managed Private Network netmask',
                  args: [{ name: 'netmask' }],
                },
                {
                  name: ['--start-ip'],
                  description:
                    'managed Private Network range start IP address',
                  args: [{ name: 'start-ip' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'Private Network zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['sg', 'security-group'],
          description: 'Security Groups management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create a Security Group',
              options: [
                {
                  name: ['--description'],
                  description: 'Security Group description',
                  args: [{ name: 'description' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete a Security Group',
              options: [
                {
                  name: ['--delete-rules', '-r'],
                  description:
                    'delete rules before deleting the Security Group',
                },
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List Security Groups',
            },
            {
              name: ['rule'],
              description: 'Security Group rules management',
              subcommands: [
                {
                  name: ['add'],
                  description: 'Add a Security Group rule',
                  options: [
                    {
                      name: ['--description'],
                      description: 'rule description',
                      args: [{ name: 'description' }],
                    },
                    {
                      name: ['--flow'],
                      description:
                        'rule network flow direction (ingress|egress)',
                      args: [{ name: 'flow', default: 'ingress' }],
                    },
                    {
                      name: ['--icmp-code'],
                      description: 'rule ICMP code',
                      args: [{ name: 'icmp-code', default: '0' }],
                    },
                    {
                      name: ['--icmp-type'],
                      description: 'rule ICMP type',
                      args: [{ name: 'icmp-type', default: '0' }],
                    },
                    {
                      name: ['--network'],
                      description:
                        'rule target network address (in CIDR format)',
                      args: [{ name: 'network' }],
                    },
                    {
                      name: ['--port'],
                      description:
                        'rule network port (format: PORT|START-END)',
                      args: [{ name: 'port' }],
                    },
                    {
                      name: ['--protocol'],
                      description: 'rule network protocol',
                      args: [{ name: 'protocol', default: 'tcp' }],
                    },
                    {
                      name: ['--security-group'],
                      description: 'rule target Security Group NAME|ID',
                      args: [{ name: 'security-group' }],
                    },
                  ],
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Security Group rule',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                  ],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Security Group details',
            },
            {
              name: ['source'],
              description: 'Security Group external sources management',
              subcommands: [
                {
                  name: ['add'],
                  description:
                    'Add an external source to a Security Group',
                },
                {
                  name: ['rm', 'remove'],
                  description:
                    'Remove an external source from a Security Group',
                },
              ],
            },
          ],
        },
        {
          name: ['sks'],
          description: 'Scalable Kubernetes Service management',
          subcommands: [
            {
              name: ['authority-cert'],
              description:
                'Retrieve an authority certificate for an SKS cluster',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['add', 'create'],
              description: 'Create an SKS cluster',
              options: [
                {
                  name: ['--auto-upgrade'],
                  description:
                    'enable automatic upgrading of the SKS cluster control plane Kubernetes version',
                },
                {
                  name: ['--cni'],
                  description:
                    "CNI plugin to deploy. e.g. 'calico', or 'cilium'",
                  args: [{ name: 'cni', default: 'calico' }],
                },
                {
                  name: ['--description'],
                  description: 'SKS cluster description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--kubernetes-version'],
                  description:
                    'SKS cluster control plane Kubernetes version',
                  args: [
                    { name: 'kubernetes-version', default: 'latest' },
                  ],
                },
                {
                  name: ['--label'],
                  description: 'SKS cluster label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--no-cni'],
                  description:
                    'do not deploy a default Container Network Interface plugin in the cluster control plane',
                },
                {
                  name: ['--no-exoscale-ccm'],
                  description:
                    'do not deploy the Exoscale Cloud Controller Manager in the cluster control plane',
                },
                {
                  name: ['--no-metrics-server'],
                  description:
                    'do not deploy the Kubernetes Metrics Server in the cluster control plane',
                },
                {
                  name: ['--nodepool-anti-affinity-group'],
                  description:
                    'default Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'nodepool-anti-affinity-group' }],
                },
                {
                  name: ['--nodepool-deploy-target'],
                  description: 'default Nodepool Deploy Target NAME|ID',
                  args: [{ name: 'nodepool-deploy-target' }],
                },
                {
                  name: ['--nodepool-description'],
                  description: 'default Nodepool description',
                  args: [{ name: 'nodepool-description' }],
                },
                {
                  name: ['--nodepool-disk-size'],
                  description:
                    'default Nodepool Compute instances disk size',
                  args: [{ name: 'nodepool-disk-size', default: '50' }],
                },
                {
                  name: ['--nodepool-instance-prefix'],
                  description:
                    'string to prefix default Nodepool member names with',
                  args: [{ name: 'nodepool-instance-prefix' }],
                },
                {
                  name: ['--nodepool-instance-type'],
                  description:
                    'default Nodepool Compute instances type',
                  args: [
                    {
                      name: 'nodepool-instance-type',
                      default: 'medium',
                    },
                  ],
                },
                {
                  name: ['--nodepool-label'],
                  description:
                    'default Nodepool label (format: key=value)',
                  args: [{ name: 'nodepool-label' }],
                },
                {
                  name: ['--nodepool-name'],
                  description: 'default Nodepool name',
                  args: [{ name: 'nodepool-name' }],
                },
                {
                  name: ['--nodepool-private-network'],
                  description:
                    'default Nodepool Private Network NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'nodepool-private-network' }],
                },
                {
                  name: ['--nodepool-security-group'],
                  description:
                    'default Nodepool Security Group NAME|ID (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'nodepool-security-group' }],
                },
                {
                  name: ['--nodepool-size'],
                  description:
                    'default Nodepool size. If 0, no default Nodepool will be added to the cluster.',
                  args: [{ name: 'nodepool-size', default: '0' }],
                },
                {
                  name: ['--nodepool-taint'],
                  description:
                    'Kubernetes taint to apply to default Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'nodepool-taint' }],
                },
                {
                  name: ['--oidc-client-id'],
                  description: 'OpenID client ID',
                  args: [{ name: 'oidc-client-id' }],
                },
                {
                  name: ['--oidc-groups-claim'],
                  description:
                    "OpenID JWT claim to use as the user's group",
                  args: [{ name: 'oidc-groups-claim' }],
                },
                {
                  name: ['--oidc-groups-prefix'],
                  description:
                    'OpenID prefix prepended to group claims',
                  args: [{ name: 'oidc-groups-prefix' }],
                },
                {
                  name: ['--oidc-issuer-url'],
                  description: 'OpenID provider URL',
                  args: [{ name: 'oidc-issuer-url' }],
                },
                {
                  name: ['--oidc-required-claim'],
                  description:
                    'OpenID token required claim (format: key=value)',
                  args: [{ name: 'oidc-required-claim' }],
                },
                {
                  name: ['--oidc-username-claim'],
                  description:
                    'OpenID JWT claim to use as the user name',
                  args: [{ name: 'oidc-username-claim' }],
                },
                {
                  name: ['--oidc-username-prefix'],
                  description:
                    'OpenID prefix prepended to username claims',
                  args: [{ name: 'oidc-username-prefix' }],
                },
                {
                  name: ['--service-level'],
                  description:
                    'SKS cluster control plane service level (starter|pro)',
                  args: [{ name: 'service-level', default: 'pro' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete an SKS cluster',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--nodepools', '-n'],
                  description:
                    'delete existing Nodepools before deleting the SKS cluster',
                },
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['dr', 'deprecated-resources'],
              description:
                'List resources that will be deprecated in a futur release of Kubernetes for an SKS cluster',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['kc', 'kubeconfig'],
              description:
                'Generate a Kubernetes kubeconfig file for an SKS cluster',
              options: [
                {
                  name: ['--exec-credential', '-x'],
                  description:
                    'output an ExecCredential object to use with a kubeconfig user.exec mode',
                },
                {
                  name: ['--group', '-g'],
                  description:
                    'client certificate group. Can be specified multiple times. Defaults to system:masters',
                  isRepeatable: true,
                  args: [{ name: 'group' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'client certificate validity duration in seconds',
                  args: [{ name: 'ttl', default: '0' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List SKS clusters',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['np', 'nodepool'],
              description: 'Manage SKS cluster Nodepools',
              subcommands: [
                {
                  name: ['add'],
                  description: 'Add a Nodepool to an SKS cluster',
                  options: [
                    {
                      name: ['--anti-affinity-group'],
                      description:
                        'Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'anti-affinity-group' }],
                    },
                    {
                      name: ['--deploy-target'],
                      description: 'Nodepool Deploy Target NAME|ID',
                      args: [{ name: 'deploy-target' }],
                    },
                    {
                      name: ['--description'],
                      description: 'Nodepool description',
                      args: [{ name: 'description' }],
                    },
                    {
                      name: ['--disk-size'],
                      description:
                        'Nodepool Compute instances disk size',
                      args: [{ name: 'disk-size', default: '50' }],
                    },
                    {
                      name: ['--instance-prefix'],
                      description:
                        'string to prefix Nodepool member names with',
                      args: [{ name: 'instance-prefix' }],
                    },
                    {
                      name: ['--instance-type'],
                      description: 'Nodepool Compute instances type',
                      args: [
                        { name: 'instance-type', default: 'medium' },
                      ],
                    },
                    {
                      name: ['--label'],
                      description: 'Nodepool label (format: key=value)',
                      isRepeatable: true,
                      args: [{ name: 'label' }],
                    },
                    {
                      name: ['--linbit'],
                      description:
                        'Create nodes with non-stadard partitioning for Linstor',
                    },
                    {
                      name: ['--private-network'],
                      description:
                        'Nodepool Private Network NAME|ID (can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'private-network' }],
                    },
                    {
                      name: ['--security-group'],
                      description:
                        'Nodepool Security Group NAME|ID (can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'security-group' }],
                    },
                    {
                      name: ['--size'],
                      description: 'Nodepool size',
                      args: [{ name: 'size', default: '2' }],
                    },
                    {
                      name: ['--taint'],
                      description:
                        'Kubernetes taint to apply to Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'taint' }],
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'SKS cluster zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete an SKS cluster Nodepool',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'SKS cluster zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['evict'],
                  description: 'Evict SKS cluster Nodepool members',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'SKS cluster zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['ls', 'list'],
                  description: 'List SKS cluster Nodepools',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'zone to filter results to',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['scale'],
                  description: 'Scale an SKS cluster Nodepool size',
                  options: [
                    {
                      name: ['--force', '-f'],
                      description: "don't prompt for confirmation",
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'SKS cluster zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an SKS cluster Nodepool details',
                  options: [
                    {
                      name: ['--zone', '-z'],
                      description: 'SKS cluster zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
                {
                  name: ['update'],
                  description: 'Update an SKS cluster Nodepool',
                  options: [
                    {
                      name: ['--anti-affinity-group'],
                      description:
                        'Nodepool Anti-Affinity Group NAME|ID (can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'anti-affinity-group' }],
                    },
                    {
                      name: ['--deploy-target'],
                      description: 'Nodepool Deploy Target NAME|ID',
                      args: [{ name: 'deploy-target' }],
                    },
                    {
                      name: ['--description'],
                      description: 'Nodepool description',
                      args: [{ name: 'description' }],
                    },
                    {
                      name: ['--disk-size'],
                      description:
                        'Nodepool Compute instances disk size',
                      args: [{ name: 'disk-size', default: '0' }],
                    },
                    {
                      name: ['--instance-prefix'],
                      description:
                        'string to prefix Nodepool member names with',
                      args: [{ name: 'instance-prefix' }],
                    },
                    {
                      name: ['--instance-type'],
                      description: 'Nodepool Compute instances type',
                      args: [{ name: 'instance-type' }],
                    },
                    {
                      name: ['--label'],
                      description:
                        'Nodepool label (format: KEY=VALUE, can be repeated multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'label' }],
                    },
                    {
                      name: ['--name'],
                      description: 'Nodepool name',
                      args: [{ name: 'name' }],
                    },
                    {
                      name: ['--private-network'],
                      description:
                        'Nodepool Private Network NAME|ID (can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'private-network' }],
                    },
                    {
                      name: ['--security-group'],
                      description:
                        'Nodepool Security Group NAME|ID (can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'security-group' }],
                    },
                    {
                      name: ['--taint'],
                      description:
                        'Kubernetes taint to apply to Nodepool Nodes (format: KEY=VALUE:EFFECT, can be specified multiple times)',
                      isRepeatable: true,
                      args: [{ name: 'taint' }],
                    },
                    {
                      name: ['--zone', '-z'],
                      description: 'SKS cluster zone',
                      args: [{ name: 'zone' }],
                    },
                  ],
                },
              ],
            },
            {
              name: ['rotate-ccm-credentials'],
              description:
                'Rotate the Exoscale Cloud Controller IAM credentials for an SKS cluster',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show an SKS cluster details',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update an SKS cluster',
              options: [
                {
                  name: ['--auto-upgrade'],
                  description:
                    'enable automatic upgrading of the SKS cluster control plane Kubernetes version',
                },
                {
                  name: ['--description'],
                  description: 'SKS cluster description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--label'],
                  description: 'SKS cluster label (format: key=value)',
                  args: [{ name: 'label' }],
                },
                {
                  name: ['--name'],
                  description: 'SKS cluster name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['upgrade'],
              description: 'Upgrade an SKS cluster Kubernetes version',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['upgrade-service-level'],
              description: 'Upgrade an SKS cluster service level',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
                {
                  name: ['--zone', '-z'],
                  description: 'SKS cluster zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
            {
              name: ['ls', 'versions'],
              description: 'List supported SKS cluster versions',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'zone to filter results to',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['ssh-key'],
          description: 'SSH keys management',
          subcommands: [
            {
              name: ['rm', 'delete'],
              description: 'Delete an SSH key',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
              ],
            },
            { name: ['list'], description: 'List SSH keys' },
            {
              name: ['add', 'register'],
              description: 'Register an SSH key',
            },
            {
              name: ['get', 'show'],
              description: 'Show an SSH key details',
            },
          ],
        },
      ],
    },
    {
      name: ['config'],
      description: 'Exoscale CLI configuration management',
      subcommands: [
        {
          name: ['add'],
          description: 'Add a new account to configuration',
        },
        {
          name: ['del', 'delete'],
          description: 'Delete an account from configuration',
          options: [
            {
              name: ['--force', '-f'],
              description:
                'attempt to perform the operation without prompting for confirmation',
            },
          ],
        },
        {
          name: ['ls', 'list'],
          description: 'List available accounts',
        },
        {
          name: ['set'],
          description: 'Set an account as default account',
        },
        {
          name: ['get', 'show'],
          description: 'Show an account details',
        },
      ],
    },
    {
      name: ['dbaas'],
      description: 'Database as a Service management',
      subcommands: [
        {
          name: ['ca-certificate'],
          description: 'Retrieve the Database CA certificate',
          options: [
            {
              name: ['--zone', '-z'],
              description: '',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['add', 'create'],
          description: 'Create a Database Service',
          options: [
            {
              name: ['--help-kafka'],
              description:
                'show usage for flags specific to the kafka type',
            },
            {
              name: ['--help-mysql'],
              description:
                'show usage for flags specific to the mysql type',
            },
            {
              name: ['--help-opensearch'],
              description:
                'show usage for flags specific to the opensearch type',
            },
            {
              name: ['--help-pg'],
              description:
                'show usage for flags specific to the pg type',
            },
            {
              name: ['--help-redis'],
              description:
                'show usage for flags specific to the redis type',
            },
            {
              name: ['--kafka-connect-settings'],
              description:
                'Kafka Connect configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-connect-settings' }],
            },
            {
              name: ['--kafka-enable-cert-auth'],
              description:
                'enable certificate-based authentication method',
              hidden: true,
            },
            {
              name: ['--kafka-enable-kafka-connect'],
              description: 'enable Kafka Connect',
              hidden: true,
            },
            {
              name: ['--kafka-enable-kafka-rest'],
              description: 'enable Kafka REST',
              hidden: true,
            },
            {
              name: ['--kafka-enable-sasl-auth'],
              description: 'enable SASL-based authentication method',
              hidden: true,
            },
            {
              name: ['--kafka-enable-schema-registry'],
              description: 'enable Schema Registry',
              hidden: true,
            },
            {
              name: ['--kafka-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'kafka-ip-filter' }],
            },
            {
              name: ['--kafka-rest-settings'],
              description:
                'Kafka REST configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-rest-settings' }],
            },
            {
              name: ['--kafka-schema-registry-settings'],
              description:
                'Schema Registry configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-schema-registry-settings' }],
            },
            {
              name: ['--kafka-settings'],
              description: 'Kafka configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-settings' }],
            },
            {
              name: ['--kafka-version'],
              description: 'Kafka major version',
              hidden: true,
              args: [{ name: 'kafka-version' }],
            },
            {
              name: ['--maintenance-dow'],
              description:
                'automated Database Service maintenance day-of-week',
              args: [{ name: 'maintenance-dow' }],
            },
            {
              name: ['--maintenance-time'],
              description:
                'automated Database Service maintenance time (format HH:MM:SS)',
              args: [{ name: 'maintenance-time' }],
            },
            {
              name: ['--mysql-admin-password'],
              description: 'custom password for admin user',
              hidden: true,
              args: [{ name: 'mysql-admin-password' }],
            },
            {
              name: ['--mysql-admin-username'],
              description: 'custom username for admin user',
              hidden: true,
              args: [{ name: 'mysql-admin-username' }],
            },
            {
              name: ['--mysql-backup-schedule'],
              description: 'automated backup schedule (format: HH:MM)',
              hidden: true,
              args: [{ name: 'mysql-backup-schedule' }],
            },
            {
              name: ['--mysql-binlog-retention-period'],
              description:
                'the minimum amount of time in seconds to keep binlog entries before deletion',
              hidden: true,
              args: [
                { name: 'mysql-binlog-retention-period', default: '0' },
              ],
            },
            {
              name: ['--mysql-fork-from'],
              description: 'name of a Database Service to fork from',
              hidden: true,
              args: [{ name: 'mysql-fork-from' }],
            },
            {
              name: ['--mysql-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'mysql-ip-filter' }],
            },
            {
              name: ['--mysql-migration-dbname'],
              description:
                'database name for bootstrapping the initial connection',
              hidden: true,
              args: [{ name: 'mysql-migration-dbname' }],
            },
            {
              name: ['--mysql-migration-host'],
              description:
                'hostname or IP address of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'mysql-migration-host' }],
            },
            {
              name: ['--mysql-migration-ignore-dbs'],
              description:
                'list of databases which should be ignored during migration',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'mysql-migration-ignore-dbs' }],
            },
            {
              name: ['--mysql-migration-method'],
              description:
                'migration method to be used ("dump" or "replication")',
              hidden: true,
              args: [{ name: 'mysql-migration-method' }],
            },
            {
              name: ['--mysql-migration-password'],
              description:
                'password for authenticating to the source server',
              hidden: true,
              args: [{ name: 'mysql-migration-password' }],
            },
            {
              name: ['--mysql-migration-port'],
              description:
                'port number of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'mysql-migration-port', default: '0' }],
            },
            {
              name: ['--mysql-migration-ssl'],
              description: 'connect to the source server using SSL',
              hidden: true,
            },
            {
              name: ['--mysql-migration-username'],
              description:
                'username for authenticating to the source server',
              hidden: true,
              args: [{ name: 'mysql-migration-username' }],
            },
            {
              name: ['--mysql-recovery-backup-time'],
              description:
                'the timestamp of the backup to restore when forking from a Database Service',
              hidden: true,
              args: [{ name: 'mysql-recovery-backup-time' }],
            },
            {
              name: ['--mysql-settings'],
              description: 'MySQL configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'mysql-settings' }],
            },
            {
              name: ['--mysql-version'],
              description: 'MySQL major version',
              hidden: true,
              args: [{ name: 'mysql-version' }],
            },
            {
              name: ['--opensearch-dashboard-enabled'],
              description:
                'Enable or disable OpenSearch Dashboards (default: true)',
              hidden: true,
            },
            {
              name: ['--opensearch-dashboard-max-old-space-size'],
              description:
                'Memory limit in MiB for OpenSearch Dashboards. Note: The memory reserved by OpenSearch Dashboards is not available for OpenSearch. (default: 128)',
              hidden: true,
              args: [
                {
                  name: 'opensearch-dashboard-max-old-space-size',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-dashboard-request-timeout'],
              description:
                'Timeout in milliseconds for requests made by OpenSearch Dashboards towards OpenSearch (default: 30000)',
              hidden: true,
              args: [
                {
                  name: 'opensearch-dashboard-request-timeout',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-fork-from-service'],
              description: 'Service name',
              hidden: true,
              args: [{ name: 'opensearch-fork-from-service' }],
            },
            {
              name: ['--opensearch-index-patterns'],
              description:
                'JSON Array of index patterns (https://openapi-v2.exoscale.com/#operation-get-dbaas-service-opensearch-200-index-patterns)',
              hidden: true,
              args: [{ name: 'opensearch-index-patterns' }],
            },
            {
              name: [
                '--opensearch-index-template-mapping-nested-objects-limit',
              ],
              description:
                'The maximum number of nested cli-flag objects that a single document can contain across all nested types. Default is 10000.',
              hidden: true,
              args: [
                {
                  name: 'opensearch-index-template-mapping-nested-objects-limit',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-index-template-number-of-replicas'],
              description:
                'The number of replicas each primary shard has.',
              hidden: true,
              args: [
                {
                  name: 'opensearch-index-template-number-of-replicas',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-index-template-number-of-shards'],
              description:
                'The number of primary shards that an index should have.',
              hidden: true,
              args: [
                {
                  name: 'opensearch-index-template-number-of-shards',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-ip-filter'],
              description:
                'Allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'opensearch-ip-filter' }],
            },
            {
              name: ['--opensearch-keep-index-refresh-interval'],
              description:
                'index.refresh_interval is reset to default value for every index to be sure that indices are always visible to search. Set to true disable this.',
              hidden: true,
            },
            {
              name: ['--opensearch-max-index-count'],
              description:
                'Maximum number of indexes to keep before deleting the oldest one',
              hidden: true,
              args: [
                { name: 'opensearch-max-index-count', default: '0' },
              ],
            },
            {
              name: ['--opensearch-recovery-backup-name'],
              description:
                'Name of a backup to recover from for services that support backup names',
              hidden: true,
              args: [{ name: 'opensearch-recovery-backup-name' }],
            },
            {
              name: ['--opensearch-settings'],
              description: 'OpenSearch-specific settings (JSON)',
              hidden: true,
              args: [{ name: 'opensearch-settings' }],
            },
            {
              name: ['--opensearch-version'],
              description: 'OpenSearch major version',
              hidden: true,
              args: [{ name: 'opensearch-version' }],
            },
            {
              name: ['--pg-admin-password'],
              description: 'custom password for admin user',
              hidden: true,
              args: [{ name: 'pg-admin-password' }],
            },
            {
              name: ['--pg-admin-username'],
              description: 'custom username for admin user',
              hidden: true,
              args: [{ name: 'pg-admin-username' }],
            },
            {
              name: ['--pg-backup-schedule'],
              description: 'automated backup schedule (format: HH:MM)',
              hidden: true,
              args: [{ name: 'pg-backup-schedule' }],
            },
            {
              name: ['--pg-bouncer-settings'],
              description:
                'PgBouncer configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'pg-bouncer-settings' }],
            },
            {
              name: ['--pg-fork-from'],
              description: 'name of a Database Service to fork from',
              hidden: true,
              args: [{ name: 'pg-fork-from' }],
            },
            {
              name: ['--pg-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'pg-ip-filter' }],
            },
            {
              name: ['--pg-lookout-settings'],
              description:
                'pglookout configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'pg-lookout-settings' }],
            },
            {
              name: ['--pg-migration-dbname'],
              description:
                'database name for bootstrapping the initial connection',
              hidden: true,
              args: [{ name: 'pg-migration-dbname' }],
            },
            {
              name: ['--pg-migration-host'],
              description:
                'hostname or IP address of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'pg-migration-host' }],
            },
            {
              name: ['--pg-migration-ignore-dbs'],
              description:
                'list of databases which should be ignored during migration',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'pg-migration-ignore-dbs' }],
            },
            {
              name: ['--pg-migration-method'],
              description:
                'migration method to be used ("dump" or "replication")',
              hidden: true,
              args: [{ name: 'pg-migration-method' }],
            },
            {
              name: ['--pg-migration-password'],
              description:
                'password for authenticating to the source server',
              hidden: true,
              args: [{ name: 'pg-migration-password' }],
            },
            {
              name: ['--pg-migration-port'],
              description:
                'port number of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'pg-migration-port', default: '0' }],
            },
            {
              name: ['--pg-migration-ssl'],
              description: 'connect to the source server using SSL',
              hidden: true,
            },
            {
              name: ['--pg-migration-username'],
              description:
                'username for authenticating to the source server',
              hidden: true,
              args: [{ name: 'pg-migration-username' }],
            },
            {
              name: ['--pg-recovery-backup-time'],
              description:
                'the timestamp of the backup to restore when forking from a Database Service',
              hidden: true,
              args: [{ name: 'pg-recovery-backup-time' }],
            },
            {
              name: ['--pg-settings'],
              description:
                'PostgreSQL configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'pg-settings' }],
            },
            {
              name: ['--pg-version'],
              description: 'PostgreSQL major version',
              hidden: true,
              args: [{ name: 'pg-version' }],
            },
            {
              name: ['--redis-fork-from'],
              description: 'name of a Database Service to fork from',
              hidden: true,
              args: [{ name: 'redis-fork-from' }],
            },
            {
              name: ['--redis-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'redis-ip-filter' }],
            },
            {
              name: ['--redis-migration-dbname'],
              description:
                'database name for bootstrapping the initial connection',
              hidden: true,
              args: [{ name: 'redis-migration-dbname' }],
            },
            {
              name: ['--redis-migration-host'],
              description:
                'hostname or IP address of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'redis-migration-host' }],
            },
            {
              name: ['--redis-migration-ignore-dbs'],
              description:
                'list of databases which should be ignored during migration',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'redis-migration-ignore-dbs' }],
            },
            {
              name: ['--redis-migration-method'],
              description:
                'migration method to be used ("dump" or "replication")',
              hidden: true,
              args: [{ name: 'redis-migration-method' }],
            },
            {
              name: ['--redis-migration-password'],
              description:
                'password for authenticating to the source server',
              hidden: true,
              args: [{ name: 'redis-migration-password' }],
            },
            {
              name: ['--redis-migration-port'],
              description:
                'port number of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'redis-migration-port', default: '0' }],
            },
            {
              name: ['--redis-migration-ssl'],
              description: 'connect to the source server using SSL',
              hidden: true,
            },
            {
              name: ['--redis-migration-username'],
              description:
                'username for authenticating to the source server',
              hidden: true,
              args: [{ name: 'redis-migration-username' }],
            },
            {
              name: ['--redis-recovery-backup-name'],
              description:
                'the name of the backup to restore when forking from a Database Service',
              hidden: true,
              args: [{ name: 'redis-recovery-backup-name' }],
            },
            {
              name: ['--redis-settings'],
              description: 'Redis configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'redis-settings' }],
            },
            {
              name: ['--termination-protection'],
              description:
                'enable Database Service termination protection; set --termination-protection=false to disable',
            },
            {
              name: ['--zone', '-z'],
              description: 'Database Service zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['rm', 'delete'],
          description: 'Delete a Database Service',
          options: [
            {
              name: ['--force', '-f'],
              description: "don't prompt for confirmation",
            },
            {
              name: ['--zone', '-z'],
              description: 'Database Service zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['ls', 'list'],
          description: 'List Database Services',
          options: [
            {
              name: ['--zone', '-z'],
              description: 'zone to filter results to',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['get', 'logs'],
          description: 'Query a Database Service logs',
          options: [
            {
              name: ['--limit', '-l'],
              description: 'number of log messages to retrieve',
              args: [{ name: 'limit', default: '10' }],
            },
            {
              name: ['--offset', '-o'],
              description:
                'opaque offset identifier (can be found in the JSON output of the command)',
              args: [{ name: 'offset' }],
            },
            {
              name: ['--sort'],
              description: 'log messages sorting order (asc|desc)',
              args: [{ name: 'sort', default: 'desc' }],
            },
            {
              name: ['--zone', '-z'],
              description: 'Database Service zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['get', 'metrics'],
          description: 'Query a Database Service metrics over time',
          options: [
            {
              name: ['--period'],
              description: 'metrics time period to retrieve',
              args: [{ name: 'period', default: 'hour' }],
            },
            {
              name: ['--zone', '-z'],
              description: 'Database Service zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['c', 'migration'],
          description: 'migration status/check',
          subcommands: [
            {
              name: ['status'],
              description: 'Migration status of a Database',
              options: [
                {
                  name: ['--zone', '-z'],
                  description: 'Database Service zone',
                  args: [{ name: 'zone' }],
                },
              ],
            },
          ],
        },
        {
          name: ['get', 'show'],
          description: 'Show a Database Service details',
          options: [
            {
              name: ['--backups'],
              description: 'show Database Service backups',
            },
            {
              name: ['--notifications'],
              description: 'show Database Service notifications',
            },
            {
              name: ['--settings'],
              description:
                'show Database Service settings (see "exo dbaas type show --help" for supported settings)',
              args: [{ name: 'settings' }],
            },
            {
              name: ['--uri'],
              description: 'show Database Service connection URI',
            },
            {
              name: ['--zone', '-z'],
              description: 'Database Service zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['type'],
          description: 'Database Services types management',
          subcommands: [
            {
              name: ['list'],
              description: 'List Database Service types',
            },
            {
              name: ['get', 'show'],
              description: 'Show a Database Service type details',
              options: [
                {
                  name: ['--backup-config'],
                  description:
                    'show backup configuration for the Database Service type and Plan',
                  args: [{ name: 'backup-config' }],
                },
                {
                  name: ['--plans'],
                  description:
                    'list plans offered for the Database Service type',
                },
                {
                  name: ['--settings'],
                  description:
                    'show settings supported by the Database Service type',
                  args: [{ name: 'settings' }],
                },
              ],
            },
          ],
        },
        {
          name: ['update'],
          description: 'Update Database Service',
          options: [
            {
              name: ['--help-kafka'],
              description:
                'show usage for flags specific to the kafka type',
            },
            {
              name: ['--help-mysql'],
              description:
                'show usage for flags specific to the mysql type',
            },
            {
              name: ['--help-opensearch'],
              description:
                'show usage for flags specific to the opensearch type',
            },
            {
              name: ['--help-pg'],
              description:
                'show usage for flags specific to the pg type',
            },
            {
              name: ['--help-redis'],
              description:
                'show usage for flags specific to the redis type',
            },
            {
              name: ['--kafka-connect-settings'],
              description:
                'Kafka Connect configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-connect-settings' }],
            },
            {
              name: ['--kafka-enable-cert-auth'],
              description:
                'enable certificate-based authentication method',
              hidden: true,
            },
            {
              name: ['--kafka-enable-kafka-connect'],
              description: 'enable Kafka Connect',
              hidden: true,
            },
            {
              name: ['--kafka-enable-kafka-rest'],
              description: 'enable Kafka REST',
              hidden: true,
            },
            {
              name: ['--kafka-enable-sasl-auth'],
              description: 'enable SASL-based authentication method',
              hidden: true,
            },
            {
              name: ['--kafka-enable-schema-registry'],
              description: 'enable Schema Registry',
              hidden: true,
            },
            {
              name: ['--kafka-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'kafka-ip-filter' }],
            },
            {
              name: ['--kafka-rest-settings'],
              description:
                'Kafka REST configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-rest-settings' }],
            },
            {
              name: ['--kafka-schema-registry-settings'],
              description:
                'Schema Registry configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-schema-registry-settings' }],
            },
            {
              name: ['--kafka-settings'],
              description: 'Kafka configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'kafka-settings' }],
            },
            {
              name: ['--maintenance-dow'],
              description:
                'automated Database Service maintenance day-of-week',
              args: [{ name: 'maintenance-dow' }],
            },
            {
              name: ['--maintenance-time'],
              description:
                'automated Database Service maintenance time (format HH:MM:SS)',
              args: [{ name: 'maintenance-time' }],
            },
            {
              name: ['--mysql-backup-schedule'],
              description: 'automated backup schedule (format: HH:MM)',
              hidden: true,
              args: [{ name: 'mysql-backup-schedule' }],
            },
            {
              name: ['--mysql-binlog-retention-period'],
              description:
                'the minimum amount of time in seconds to keep binlog entries before deletion',
              hidden: true,
              args: [
                { name: 'mysql-binlog-retention-period', default: '0' },
              ],
            },
            {
              name: ['--mysql-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'mysql-ip-filter' }],
            },
            {
              name: ['--mysql-migration-dbname'],
              description:
                'database name for bootstrapping the initial connection',
              hidden: true,
              args: [{ name: 'mysql-migration-dbname' }],
            },
            {
              name: ['--mysql-migration-host'],
              description:
                'hostname or IP address of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'mysql-migration-host' }],
            },
            {
              name: ['--mysql-migration-ignore-dbs'],
              description:
                'list of databases which should be ignored during migration',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'mysql-migration-ignore-dbs' }],
            },
            {
              name: ['--mysql-migration-method'],
              description:
                'migration method to be used ("dump" or "replication")',
              hidden: true,
              args: [{ name: 'mysql-migration-method' }],
            },
            {
              name: ['--mysql-migration-password'],
              description:
                'password for authenticating to the source server',
              hidden: true,
              args: [{ name: 'mysql-migration-password' }],
            },
            {
              name: ['--mysql-migration-port'],
              description:
                'port number of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'mysql-migration-port', default: '0' }],
            },
            {
              name: ['--mysql-migration-ssl'],
              description: 'connect to the source server using SSL',
              hidden: true,
            },
            {
              name: ['--mysql-migration-username'],
              description:
                'username for authenticating to the source server',
              hidden: true,
              args: [{ name: 'mysql-migration-username' }],
            },
            {
              name: ['--mysql-settings'],
              description: 'MySQL configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'mysql-settings' }],
            },
            {
              name: ['--opensearch-dashboard-enabled'],
              description:
                'Enable or disable OpenSearch Dashboards (default: true)',
              hidden: true,
            },
            {
              name: ['--opensearch-dashboard-max-old-space-size'],
              description:
                'Memory limit in MiB for OpenSearch Dashboards. Note: The memory reserved by OpenSearch Dashboards is not available for OpenSearch. (default: 128)',
              hidden: true,
              args: [
                {
                  name: 'opensearch-dashboard-max-old-space-size',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-dashboard-request-timeout'],
              description:
                'Timeout in milliseconds for requests made by OpenSearch Dashboards towards OpenSearch (default: 30000)',
              hidden: true,
              args: [
                {
                  name: 'opensearch-dashboard-request-timeout',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-index-patterns'],
              description:
                'JSON Array of index patterns (https://openapi-v2.exoscale.com/#operation-get-dbaas-service-opensearch-200-index-patterns)',
              hidden: true,
              args: [{ name: 'opensearch-index-patterns' }],
            },
            {
              name: [
                '--opensearch-index-template-mapping-nested-objects-limit',
              ],
              description:
                'The maximum number of nested cli-flag objects that a single document can contain across all nested types. Default is 10000.',
              hidden: true,
              args: [
                {
                  name: 'opensearch-index-template-mapping-nested-objects-limit',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-index-template-number-of-replicas'],
              description:
                'The number of replicas each primary shard has.',
              hidden: true,
              args: [
                {
                  name: 'opensearch-index-template-number-of-replicas',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-index-template-number-of-shards'],
              description:
                'The number of primary shards that an index should have.',
              hidden: true,
              args: [
                {
                  name: 'opensearch-index-template-number-of-shards',
                  default: '0',
                },
              ],
            },
            {
              name: ['--opensearch-ip-filter'],
              description:
                'Allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'opensearch-ip-filter' }],
            },
            {
              name: ['--opensearch-keep-index-refresh-interval'],
              description:
                'index.refresh_interval is reset to default value for every index to be sure that indices are always visible to search. Set to true disable this.',
              hidden: true,
            },
            {
              name: ['--opensearch-max-index-count'],
              description:
                'Maximum number of indexes to keep before deleting the oldest one',
              hidden: true,
              args: [
                { name: 'opensearch-max-index-count', default: '0' },
              ],
            },
            {
              name: ['--opensearch-settings'],
              description: 'OpenSearch-specific settings (JSON)',
              hidden: true,
              args: [{ name: 'opensearch-settings' }],
            },
            {
              name: ['--pg-backup-schedule'],
              description: 'automated backup schedule (format: HH:MM)',
              hidden: true,
              args: [{ name: 'pg-backup-schedule' }],
            },
            {
              name: ['--pg-bouncer-settings'],
              description:
                'PgBouncer configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'pg-bouncer-settings' }],
            },
            {
              name: ['--pg-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'pg-ip-filter' }],
            },
            {
              name: ['--pg-lookout-settings'],
              description:
                'pglookout configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'pg-lookout-settings' }],
            },
            {
              name: ['--pg-migration-dbname'],
              description:
                'database name for bootstrapping the initial connection',
              hidden: true,
              args: [{ name: 'pg-migration-dbname' }],
            },
            {
              name: ['--pg-migration-host'],
              description:
                'hostname or IP address of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'pg-migration-host' }],
            },
            {
              name: ['--pg-migration-ignore-dbs'],
              description:
                'list of databases which should be ignored during migration',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'pg-migration-ignore-dbs' }],
            },
            {
              name: ['--pg-migration-method'],
              description:
                'migration method to be used ("dump" or "replication")',
              hidden: true,
              args: [{ name: 'pg-migration-method' }],
            },
            {
              name: ['--pg-migration-password'],
              description:
                'password for authenticating to the source server',
              hidden: true,
              args: [{ name: 'pg-migration-password' }],
            },
            {
              name: ['--pg-migration-port'],
              description:
                'port number of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'pg-migration-port', default: '0' }],
            },
            {
              name: ['--pg-migration-ssl'],
              description: 'connect to the source server using SSL',
              hidden: true,
            },
            {
              name: ['--pg-migration-username'],
              description:
                'username for authenticating to the source server',
              hidden: true,
              args: [{ name: 'pg-migration-username' }],
            },
            {
              name: ['--pg-settings'],
              description:
                'PostgreSQL configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'pg-settings' }],
            },
            {
              name: ['--plan'],
              description: 'Database Service plan',
              args: [{ name: 'plan' }],
            },
            {
              name: ['--redis-ip-filter'],
              description:
                'allow incoming connections from CIDR address block',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'redis-ip-filter' }],
            },
            {
              name: ['--redis-migration-dbname'],
              description:
                'database name for bootstrapping the initial connection',
              hidden: true,
              args: [{ name: 'redis-migration-dbname' }],
            },
            {
              name: ['--redis-migration-host'],
              description:
                'hostname or IP address of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'redis-migration-host' }],
            },
            {
              name: ['--redis-migration-ignore-dbs'],
              description:
                'list of databases which should be ignored during migration',
              hidden: true,
              isRepeatable: true,
              args: [{ name: 'redis-migration-ignore-dbs' }],
            },
            {
              name: ['--redis-migration-method'],
              description:
                'migration method to be used ("dump" or "replication")',
              hidden: true,
              args: [{ name: 'redis-migration-method' }],
            },
            {
              name: ['--redis-migration-password'],
              description:
                'password for authenticating to the source server',
              hidden: true,
              args: [{ name: 'redis-migration-password' }],
            },
            {
              name: ['--redis-migration-port'],
              description:
                'port number of the source server where to migrate data from',
              hidden: true,
              args: [{ name: 'redis-migration-port', default: '0' }],
            },
            {
              name: ['--redis-migration-ssl'],
              description: 'connect to the source server using SSL',
              hidden: true,
            },
            {
              name: ['--redis-migration-username'],
              description:
                'username for authenticating to the source server',
              hidden: true,
              args: [{ name: 'redis-migration-username' }],
            },
            {
              name: ['--redis-settings'],
              description: 'Redis configuration settings (JSON format)',
              hidden: true,
              args: [{ name: 'redis-settings' }],
            },
            {
              name: ['--termination-protection'],
              description:
                'enable Database Service termination protection; set --termination-protection=false to disable',
            },
            {
              name: ['--zone', '-z'],
              description: 'Database Service zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
      ],
    },
    {
      name: ['dns'],
      description:
        'DNS cmd lets you host your zones and manage records',
      subcommands: [
        {
          name: ['add'],
          description: 'Add record to domain',
          subcommands: [
            {
              name: ['A'],
              description: 'Add A record type to a domain',
              options: [
                {
                  name: ['--address', '-a'],
                  description: 'Example: 127.0.0.1',
                  args: [{ name: 'address' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['AAAA'],
              description: 'Add AAAA record type to a domain',
              options: [
                {
                  name: ['--address', '-a'],
                  description:
                    'Example: 2001:0db8:85a3:0000:0000:EA75:1337:BEEF',
                  args: [{ name: 'address' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['ALIAS'],
              description: 'Add ALIAS record type to a domain',
              options: [
                {
                  name: ['--alias', '-a'],
                  description:
                    'Alias for: Example: some-other-site.com',
                  args: [{ name: 'alias' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['CAA'],
              description: 'Add CAA record type to a domain',
              options: [
                {
                  name: ['--flag', '-f'],
                  description: 'An unsigned integer between 0-255.',
                  args: [{ name: 'flag', default: '0' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--tag'],
                  description:
                    'CAA tag "KEY,VALUE", available tags: (issue|issuewild|iodef)',
                  isRepeatable: true,
                  args: [{ name: 'tag' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['CNAME'],
              description: 'Add CNAME record type to a domain',
              options: [
                {
                  name: ['--alias', '-a'],
                  description:
                    'Alias for: Example: some-other-site.com',
                  args: [{ name: 'alias' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'You may use the * wildcard here.',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['HINFO'],
              description: 'Add HINFO record type to a domain',
              options: [
                {
                  name: ['--cpu', '-c'],
                  description: 'Example: IBM-PC/AT',
                  args: [{ name: 'cpu' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--os', '-o'],
                  description:
                    'The operating system of the machine, example: Linux',
                  args: [{ name: 'os' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['MX'],
              description: 'Add MX record type to a domain',
              options: [
                {
                  name: ['--mail-server-host', '-m'],
                  description: 'Example: mail-server.example.com',
                  args: [{ name: 'mail-server-host' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    'Leave this blank to create a record for DOMAIN',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description:
                    'Common values are for example 1, 5 or 10',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['NAPTR'],
              description: 'Add NAPTR record type to a domain',
              options: [
                {
                  name: ['--a'],
                  description:
                    'Flag indicates the next lookup is for an A or AAAA record.',
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--order', '-o'],
                  description:
                    'Used to determine the processing order, lowest first.',
                  args: [{ name: 'order', default: '0' }],
                },
                {
                  name: ['--p'],
                  description:
                    'Flag indicates that processing should continue in a protocol-specific fashion.',
                },
                {
                  name: ['--preference'],
                  description:
                    "Used to give weight to records with the same value in the 'order' field, low to high.",
                  args: [{ name: 'preference', default: '0' }],
                },
                {
                  name: ['--regex'],
                  description: 'The substitution expression.',
                  args: [{ name: 'regex' }],
                },
                {
                  name: ['--replacement'],
                  description:
                    'The next record to look up, which must be a fully-qualified domain name.',
                  args: [{ name: 'replacement' }],
                },
                {
                  name: ['--s'],
                  description:
                    'Flag indicates the next lookup is for an SRV.',
                },
                {
                  name: ['--service'],
                  description: 'Service',
                  args: [{ name: 'service' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
                {
                  name: ['--u'],
                  description:
                    'Flag indicates the next record is the output of the regular expression as a URI.',
                },
              ],
            },
            {
              name: ['NS'],
              description: 'Add NS record type to a domain',
              options: [
                {
                  name: ['--name', '-n'],
                  description: 'You may use the * wildcard here.',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--name-server', '-s'],
                  description: "Example: 'ns1.example.com'",
                  args: [{ name: 'name-server' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['POOL'],
              description: 'Add POOL record type to a domain',
              options: [
                {
                  name: ['--alias', '-a'],
                  description:
                    "Alias for: Example: 'some-other-site.com'",
                  args: [{ name: 'alias' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'You may use the * wildcard here.',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['SRV'],
              description: 'Add SRV record type to a domain',
              options: [
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--port', '-P'],
                  description:
                    "The 'TCP' or 'UDP' port on which the service is found.",
                  args: [{ name: 'port' }],
                },
                {
                  name: ['--priority'],
                  description: 'Priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--protocol', '-p'],
                  description: "This will usually be 'TCP' or 'UDP'.",
                  args: [{ name: 'protocol' }],
                },
                {
                  name: ['--symbolic-name', '-s'],
                  description:
                    "This will be a symbolic name for the service, like 'sip'. It might also be called Service at other DNS providers.",
                  args: [{ name: 'symbolic-name' }],
                },
                {
                  name: ['--target'],
                  description:
                    'The canonical hostname of the machine providing the service.',
                  args: [{ name: 'target' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
                {
                  name: ['--weight', '-w'],
                  description:
                    "A relative weight for 'SRV' records with the same priority.",
                  args: [{ name: 'weight', default: '0' }],
                },
              ],
            },
            {
              name: ['SSHFP'],
              description: 'Add SSHFP record type to a domain',
              options: [
                {
                  name: ['--algorithm', '-a'],
                  description:
                    'RSA(1) | DSA(2) | ECDSA(3) | ED25519(4)',
                  args: [{ name: 'algorithm', default: '0' }],
                },
                {
                  name: ['--fingerprint', '-f'],
                  description: 'Fingerprint',
                  args: [{ name: 'fingerprint' }],
                },
                {
                  name: ['--fingerprint-type'],
                  description: 'SHA1(1) | SHA256(2)',
                  args: [{ name: 'fingerprint-type', default: '0' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['TXT'],
              description: 'Add TXT record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Content record',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
            {
              name: ['URL'],
              description: 'Add URL record type to a domain',
              options: [
                {
                  name: ['--destination-url', '-d'],
                  description: 'Example: https://www.example.com',
                  args: [{ name: 'destination-url' }],
                },
                {
                  name: ['--name', '-n'],
                  description:
                    "Leave this blank to create a record for DOMAIN, You may use the '*' wildcard here.",
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description:
                    'The time in seconds to live (refresh rate) of the record.',
                  args: [{ name: 'ttl', default: '3600' }],
                },
              ],
            },
          ],
        },
        { name: ['add', 'create'], description: 'Create a domain' },
        {
          name: ['del', 'delete'],
          description: 'Delete a domain',
          options: [
            {
              name: ['--force', '-f'],
              description:
                'attempt to perform the operation without prompting for confirmation',
            },
          ],
        },
        { name: ['ls', 'list'], description: 'List domains' },
        {
          name: ['rm', 'remove'],
          description: 'Remove a domain record',
          options: [
            {
              name: ['--force', '-f'],
              description:
                'attempt to perform the operation without prompting for confirmation',
            },
          ],
        },
        {
          name: ['show'],
          description: 'Show the domain records',
          options: [
            {
              name: ['--name', '-n'],
              description: 'List records by name',
              args: [{ name: 'name' }],
            },
          ],
        },
        {
          name: ['update'],
          description: 'Update domain record',
          subcommands: [
            {
              name: ['A'],
              description: 'Update A record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['AAAA'],
              description: 'Update AAAA record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['ALIAS'],
              description: 'Update ALIAS record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['CNAME'],
              description: 'Update CNAME record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['HINFO'],
              description: 'Update HINFO record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['MX'],
              description: 'Update MX record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['NAPTR'],
              description: 'Update NAPTR record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['NS'],
              description: 'Update NS record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['POOL'],
              description: 'Update POOL record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['SPF'],
              description: 'Update SPF record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['SRV'],
              description: 'Update SRV record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['SSHFP'],
              description: 'Update SSHFP record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['TXT'],
              description: 'Update TXT record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
            {
              name: ['URL'],
              description: 'Update URL record type to a domain',
              options: [
                {
                  name: ['--content', '-c'],
                  description: 'Update Content',
                  args: [{ name: 'content' }],
                },
                {
                  name: ['--name', '-n'],
                  description: 'Update name',
                  args: [{ name: 'name' }],
                },
                {
                  name: ['--priority', '-p'],
                  description: 'Update priority',
                  args: [{ name: 'priority', default: '0' }],
                },
                {
                  name: ['--ttl', '-t'],
                  description: 'Update ttl',
                  args: [{ name: 'ttl', default: '0' }],
                },
              ],
            },
          ],
        },
      ],
    },
    {
      name: ['environment'],
      description: 'Environment variables usage',
    },
    {
      name: ['iam'],
      description: 'Identity and Access Management',
      subcommands: [
        {
          name: ['key', 'access-key'],
          description: 'IAM access keys management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create an IAM access key',
              options: [
                {
                  name: ['--operation'],
                  description:
                    'API operation to restrict the access key to. Can be repeated multiple times.',
                  isRepeatable: true,
                  args: [{ name: 'operation' }],
                },
                {
                  name: ['--resource'],
                  description:
                    'API resource to restrict the access key to (format: DOMAIN/TYPE:NAME). Can be repeated multiple times.',
                  isRepeatable: true,
                  args: [{ name: 'resource' }],
                },
                {
                  name: ['--tag'],
                  description:
                    'API operations tag to restrict the access key to. Can be repeated multiple times.',
                  isRepeatable: true,
                  args: [{ name: 'tag' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List IAM access keys',
            },
            {
              name: ['ls', 'list-operations'],
              description: 'List IAM access keys operations',
              options: [
                {
                  name: ['--mine'],
                  description:
                    'only report operations available to the IAM access key used to perform the API request',
                },
              ],
            },
            {
              name: ['add', 'revoke'],
              description: 'Revoke an IAM access key',
              options: [
                {
                  name: ['--force', '-f'],
                  description: "don't prompt for confirmation",
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show an IAM access key details',
            },
          ],
        },
      ],
    },
    {
      name: ['lab'],
      description: 'Experimental commands',
      subcommands: [
        {
          name: ['coi'],
          description: 'Deploy a Container-Optimized Instance',
          options: [
            {
              name: ['--disk-size', '-d'],
              description: 'disk size',
              args: [{ name: 'disk-size', default: '20' }],
            },
            {
              name: ['--docker-compose', '-c'],
              description:
                'Docker Compose configuration file (local path/URL)',
              args: [{ name: 'docker-compose' }],
            },
            {
              name: ['--image', '-i'],
              description: 'Docker image to run',
              args: [{ name: 'image' }],
            },
            {
              name: ['--port', '-p'],
              description:
                'publish container port(s) (format "[HOST-PORT:]CONTAINER-PORT")',
              isRepeatable: true,
              args: [{ name: 'port' }],
            },
            {
              name: ['--private-network'],
              description: 'Private Network name/ID',
              isRepeatable: true,
              args: [{ name: 'private-network' }],
            },
            {
              name: ['--security-group'],
              description: 'Security Group name/ID',
              isRepeatable: true,
              args: [{ name: 'security-group' }],
            },
            {
              name: ['--service-offering', '-o'],
              description:
                'service offering NAME (micro|tiny|small|medium|large|extra-large|huge|mega|titan|jumbo)',
              args: [{ name: 'service-offering', default: 'medium' }],
            },
            {
              name: ['--ssh-key', '-k'],
              description: 'SSH key name',
              args: [{ name: 'ssh-key' }],
            },
            {
              name: ['--zone', '-z'],
              description:
                'zone NAME|ID (ch-dk-2|ch-gva-2|at-vie-1|de-fra-1|bg-sof-1|de-muc-1)',
              args: [{ name: 'zone' }],
            },
          ],
        },
      ],
    },
    { name: ['limits'], description: 'Current account limits' },
    { name: ['output'], description: 'Output formatting usage' },
    {
      name: ['runstatus'],
      description: 'Manage your Runstat.us pages',
      subcommands: [
        {
          name: ['add', 'create'],
          description: 'Create Runstat.us page',
          options: [
            {
              name: ['--dark', '-d'],
              description: 'Enable status page dark mode',
            },
          ],
        },
        {
          name: ['del', 'delete'],
          description: 'Delete runstat.us page(s)',
        },
        {
          name: ['incident'],
          description: 'Incident management',
          subcommands: [
            {
              name: ['add'],
              description: 'Add an incident to a runstat.us page',
              options: [
                {
                  name: ['--description', '-d'],
                  description: 'incident initial event description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--services'],
                  description:
                    'List of strings with the services impacted. e.g: service1,service2,...',
                  isRepeatable: true,
                  args: [{ name: 'services' }],
                },
                {
                  name: ['--state', '-s'],
                  description:
                    'incident state (major_outage|partial_outage|degraded_performance|operational)',
                  args: [{ name: 'state' }],
                },
                {
                  name: ['--status'],
                  description:
                    'incident status (investigating|identified|monitoring)',
                  args: [{ name: 'status' }],
                },
                {
                  name: ['--title', '-t'],
                  description: 'incident title',
                  args: [{ name: 'title' }],
                },
              ],
            },
            { name: ['ls', 'list'], description: 'List incidents' },
            {
              name: ['rm', 'remove'],
              description: 'Remove incident from a runstat.us page',
            },
            {
              name: ['get', 'show'],
              description: 'Show an incident details',
            },
            {
              name: ['update'],
              description: 'update an existing incident',
              options: [
                {
                  name: ['--description', '-d'],
                  description: 'incident description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--state', '-t'],
                  description:
                    'incident state (major_outage|partial_outage|degraded_performance|operational)',
                  args: [{ name: 'state' }],
                },
                {
                  name: ['--status', '-s'],
                  description:
                    'incident status (investigating|identified|monitoring|resolved)',
                  args: [{ name: 'status' }],
                },
              ],
            },
          ],
        },
        { name: ['ls', 'list'], description: 'List runstat.us pages' },
        {
          name: ['maintenance'],
          description: 'Maintenance management',
          subcommands: [
            {
              name: ['add'],
              description: 'Add a maintenance to a runstat.us page',
              options: [
                {
                  name: ['--description', '-d'],
                  description: 'maintenance description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--end-date', '-e'],
                  description:
                    'maintenance end date in UTC format (e.g. "2016-05-31T21:11:32.378Z")',
                  args: [{ name: 'end-date' }],
                },
                {
                  name: ['--services'],
                  description:
                    'service affected by the maintenance (can be specified multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'services' }],
                },
                {
                  name: ['--start-date', '-s'],
                  description:
                    'maintenance start date in UTC format (e.g. "2016-05-31T21:11:32.378Z")',
                  args: [{ name: 'start-date' }],
                },
                {
                  name: ['--status'],
                  description:
                    'maintenance status (scheduled|in-progress|completed)',
                  args: [{ name: 'status' }],
                },
                {
                  name: ['--title', '-t'],
                  description: 'maintenance title',
                  args: [{ name: 'title' }],
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List maintenance from page(s)',
            },
            {
              name: ['rm', 'remove'],
              description: 'Remove maintenance from a runstat.us page',
            },
            {
              name: ['get', 'show'],
              description: 'Show a maintenance details',
            },
            {
              name: ['update'],
              description: 'update a maintenance',
              options: [
                {
                  name: ['--description', '-d'],
                  description: 'maintenance description',
                  args: [{ name: 'description' }],
                },
                {
                  name: ['--status', '-s'],
                  description:
                    'maintenance status (scheduled|in-progress|completed)',
                  args: [{ name: 'status' }],
                },
              ],
            },
          ],
        },
        {
          name: ['service'],
          description: 'Runstat.us service management',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create a service',
            },
            {
              name: ['del', 'delete'],
              description: 'Delete a service',
            },
            { name: ['ls', 'list'], description: 'List services' },
            {
              name: ['get', 'show'],
              description: 'Show a service details',
            },
          ],
        },
        {
          name: ['get', 'show'],
          description: 'Show a runstat.us page details',
        },
      ],
    },
    { name: ['status'], description: 'Exoscale status' },
    {
      name: ['storage'],
      description: 'Object Storage management',
      subcommands: [
        {
          name: ['cors'],
          description: 'Manage buckets CORS configuration',
          subcommands: [
            {
              name: ['add'],
              description: 'Add a CORS configuration rule to a bucket',
              options: [
                {
                  name: ['--allowed-header'],
                  description:
                    'allowed header (can be repeated multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'allowed-header' }],
                },
                {
                  name: ['--allowed-method'],
                  description:
                    'allowed method (can be repeated multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'allowed-method' }],
                },
                {
                  name: ['--allowed-origin'],
                  description:
                    'allowed origin (can be repeated multiple times)',
                  isRepeatable: true,
                  args: [{ name: 'allowed-origin' }],
                },
              ],
            },
            {
              name: ['del', 'delete'],
              description: 'Delete the CORS configuration of a bucket',
              options: [
                {
                  name: ['--force', '-f'],
                  description:
                    'attempt to perform the operation without prompting for confirmation',
                },
              ],
            },
          ],
        },
        {
          name: ['del', 'rm', 'delete'],
          description: 'Delete objects',
          options: [
            {
              name: ['--force', '-f'],
              description:
                'attempt to perform the operation without prompting for confirmation',
            },
            {
              name: ['--recursive', '-r'],
              description: 'delete objects recursively',
            },
            {
              name: ['--verbose', '-v'],
              description: 'output deleted objects',
            },
          ],
        },
        {
          name: ['get', 'download'],
          description: 'Download files from a bucket',
          options: [
            {
              name: ['--dry-run', '-n'],
              description:
                "simulate files download, don't actually do it",
            },
            {
              name: ['--force', '-f'],
              description: 'overwrite existing destination files',
            },
            {
              name: ['--recursive', '-r'],
              description: 'download prefix recursively',
            },
          ],
        },
        {
          name: ['headers'],
          description: 'Manage objects HTTP headers',
          subcommands: [
            {
              name: ['add'],
              description: 'Add HTTP headers to an object',
              options: [
                {
                  name: ['--cache-control'],
                  description: 'value for "Cache-Control" header',
                  args: [{ name: 'cache-control' }],
                },
                {
                  name: ['--content-disposition'],
                  description: 'value for "Content-Disposition" header',
                  args: [{ name: 'content-disposition' }],
                },
                {
                  name: ['--content-encoding'],
                  description: 'value for "Content-Encoding" header',
                  args: [{ name: 'content-encoding' }],
                },
                {
                  name: ['--content-language'],
                  description: 'value for "Content-Language" header',
                  args: [{ name: 'content-language' }],
                },
                {
                  name: ['--content-type'],
                  description: 'value for "Content-Type" header',
                  args: [{ name: 'content-type' }],
                },
                {
                  name: ['--expires'],
                  description: 'value for "Expires" header',
                  args: [{ name: 'expires' }],
                },
                {
                  name: ['--recursive', '-r'],
                  description:
                    'add headers recursively (with object prefix only)',
                },
              ],
            },
            {
              name: ['del', 'delete'],
              description: 'Delete HTTP headers from an object',
              options: [
                {
                  name: ['--cache-control'],
                  description: 'delete the "Cache-Control" header',
                },
                {
                  name: ['--content-disposition'],
                  description:
                    'delete the "Content-Disposition" header',
                },
                {
                  name: ['--content-encoding'],
                  description: 'delete the "Content-Encoding" header',
                },
                {
                  name: ['--content-language'],
                  description: 'delete the "Content-Language" header',
                },
                {
                  name: ['--content-type'],
                  description: 'delete the "Content-Type" header',
                },
                {
                  name: ['--expires'],
                  description: 'delete the "Expires" header',
                },
                {
                  name: ['--recursive', '-r'],
                  description:
                    'delete headers recursively (with object prefix only)',
                },
              ],
            },
          ],
        },
        {
          name: ['ls', 'list'],
          description: 'List buckets and objects',
          options: [
            {
              name: ['--recursive', '-r'],
              description: 'list bucket recursively',
            },
            {
              name: ['--stream', '-s'],
              description:
                'stream listed files instead of waiting for complete listing (useful for large buckets)',
            },
          ],
        },
        {
          name: ['create', 'mb'],
          description: 'Create a new bucket',
          options: [
            {
              name: ['--acl'],
              description:
                'canned ACL to set on bucket (private|public-read|public-read-write|authenticated-read)',
              args: [{ name: 'acl' }],
            },
            {
              name: ['--zone', '-z'],
              description: 'bucket zone',
              args: [{ name: 'zone' }],
            },
          ],
        },
        {
          name: ['meta', 'metadata'],
          description: 'Manage objects metadata',
          subcommands: [
            {
              name: ['add'],
              description: 'Add key/value metadata to an object',
              options: [
                {
                  name: ['--recursive', '-r'],
                  description:
                    'add metadata recursively (with object prefix only)',
                },
              ],
            },
            {
              name: ['del', 'delete'],
              description: 'Delete metadata from an object',
              options: [
                {
                  name: ['--recursive', '-r'],
                  description:
                    'delete metadata recursively (with object prefix only)',
                },
              ],
            },
          ],
        },
        {
          name: ['presign'],
          description: 'Generate a pre-signed URL to an object',
          options: [
            {
              name: ['--expires', '-e'],
              description:
                'expiration duration for the generated pre-signed URL (e.g. "1h45m", "30s"); supported units: "s", "m", "h"',
              args: [{ name: 'expires', default: '15m0s' }],
            },
            {
              name: ['--method', '-m'],
              description: 'pre-signed URL method (get|put)',
              args: [{ name: 'method', default: 'get' }],
            },
          ],
        },
        {
          name: ['rb'],
          description: 'Delete a bucket',
          options: [
            {
              name: ['--force', '-f'],
              description:
                'attempt to perform the operation without prompting for confirmation',
            },
            {
              name: ['--recursive', '-r'],
              description: 'empty the bucket before deleting it',
            },
          ],
        },
        {
          name: ['setacl'],
          description: 'Set a bucket/objects ACL',
          options: [
            {
              name: ['--full-control'],
              description: 'ACP Full Control grantee',
              args: [{ name: 'full-control' }],
            },
            {
              name: ['--read'],
              description: 'ACL Read grantee',
              args: [{ name: 'read' }],
            },
            {
              name: ['--read-acp'],
              description: 'ACP Read ACP grantee',
              args: [{ name: 'read-acp' }],
            },
            {
              name: ['--recursive', '-r'],
              description:
                'set ACL recursively (with object prefix only)',
            },
            {
              name: ['--write'],
              description: 'ACP Write grantee',
              args: [{ name: 'write' }],
            },
            {
              name: ['--write-acp'],
              description: 'ACP Write ACP grantee',
              args: [{ name: 'write-acp' }],
            },
          ],
        },
        { name: ['show'], description: 'Show a bucket/object details' },
        {
          name: ['put', 'upload'],
          description: 'Upload files to a bucket',
          options: [
            {
              name: ['--acl'],
              description:
                'canned ACL to set on object (private|public-read|public-read-write|authenticated-read|aws-exec-read|bucket-owner-read|bucket-owner-full-control)',
              args: [{ name: 'acl' }],
            },
            {
              name: ['--dry-run', '-n'],
              description:
                "simulate files upload, don't actually do it",
            },
            {
              name: ['--recursive', '-r'],
              description: 'upload directories recursively',
            },
          ],
        },
      ],
    },
    { name: ['version'], description: 'Print the version of exo' },
    {
      name: ['zones', 'zone'],
      description: 'List all available zones',
    },
    {
      name: ['help'],
      description: 'Help about any command',
      subcommands: [
        {
          name: ['completion'],
          description:
            'Generate the autocompletion script for the specified shell',
          subcommands: [
            {
              name: ['bash'],
              description:
                'Generate the autocompletion script for bash',
            },
            {
              name: ['fish'],
              description:
                'Generate the autocompletion script for fish',
            },
            {
              name: ['powershell'],
              description:
                'Generate the autocompletion script for powershell',
            },
            {
              name: ['zsh'],
              description: 'Generate the autocompletion script for zsh',
            },
          ],
        },
        {
          name: ['c', 'compute'],
          description: 'Compute services management',
          subcommands: [
            {
              name: ['aag', 'anti-affinity-group'],
              description: 'Anti-Affinity Groups management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create an Anti-Affinity Group',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete an Anti-Affinity Group',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Anti-Affinity Groups',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an Anti-Affinity Group details',
                },
              ],
            },
            {
              name: ['dt', 'deploy-target'],
              description: 'Compute instance Deploy Targets management',
              subcommands: [
                { name: ['list'], description: 'List Deploy Targets' },
                {
                  name: ['get', 'show'],
                  description: 'Show a Deploy Target details',
                },
              ],
            },
            {
              name: ['eip', 'elastic-ip'],
              description: 'Elastic IP addresses management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create an Elastic IP',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete an Elastic IP',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Elastic IPs',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an Elastic IP details',
                },
                {
                  name: ['update'],
                  description: 'Update an Elastic IP',
                },
              ],
            },
            {
              name: ['i', 'instance'],
              description: 'Compute instances management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create a Compute instance',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Compute instance',
                },
                {
                  name: ['eip', 'elastic-ip'],
                  description:
                    'Manage Compute instance Elastic IP addresses',
                  subcommands: [
                    {
                      name: ['attach'],
                      description:
                        'Attach an Elastic IP to a Compute instance',
                    },
                    {
                      name: ['detach'],
                      description:
                        'Detach a Compute instance from a Elastic IP',
                    },
                  ],
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Compute instances',
                },
                {
                  name: ['privnet', 'private-network'],
                  description:
                    'Manage Compute instance Private Networks',
                  subcommands: [
                    {
                      name: ['attach'],
                      description:
                        'Attach a Compute instance to a Private Network',
                    },
                    {
                      name: ['detach'],
                      description:
                        'Detach a Compute instance from a Private Network',
                    },
                    {
                      name: ['update-ip'],
                      description:
                        'Update a Compute instance Private Network IP address',
                    },
                  ],
                },
                {
                  name: ['reboot'],
                  description: 'Reboot a Compute instance',
                },
                {
                  name: ['reset'],
                  description: 'Reset a Compute instance',
                },
                {
                  name: ['resize-disk'],
                  description: 'Resize a Compute instance disk',
                },
                {
                  name: ['scale'],
                  description: 'Scale a Compute instance',
                },
                {
                  name: ['scp'],
                  description: 'SCP files to/from a Compute instance',
                },
                {
                  name: ['sg', 'security-group'],
                  description:
                    'Manage Compute instance Security Groups',
                  subcommands: [
                    {
                      name: ['add'],
                      description:
                        'Add a Compute instance to Security Groups',
                    },
                    {
                      name: ['rm', 'remove'],
                      description:
                        'Remove a Compute instance from Security Groups',
                    },
                  ],
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a Compute instance details',
                },
                {
                  name: ['snap', 'snapshot'],
                  description: 'Manage Compute instance snapshots',
                  subcommands: [
                    {
                      name: ['add', 'create'],
                      description: 'Create a Compute instance snapshot',
                    },
                    {
                      name: ['rm', 'delete'],
                      description: 'Delete a Compute instance snapshot',
                    },
                    {
                      name: ['export'],
                      description: 'Export a Compute instance snapshot',
                    },
                    {
                      name: ['list'],
                      description: 'List Compute instance snapshots',
                    },
                    {
                      name: ['revert'],
                      description:
                        'Revert a Compute instance to a snapshot',
                    },
                    {
                      name: ['get', 'show'],
                      description:
                        'Show a Compute instance snapshot details',
                    },
                  ],
                },
                {
                  name: ['ssh'],
                  description: 'Log into a Compute instance via SSH',
                },
                {
                  name: ['start'],
                  description: 'Start a Compute instance',
                },
                {
                  name: ['stop'],
                  description: 'Stop a Compute instance',
                },
                {
                  name: ['update'],
                  description: 'Update an Instance ',
                },
              ],
            },
            {
              name: ['pool', 'instance-pool'],
              description: 'Instance Pools management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create an Instance Pool',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete an Instance Pool',
                },
                {
                  name: ['evict'],
                  description: 'Evict Instance Pool members',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Instance Pools',
                },
                {
                  name: ['scale'],
                  description: 'Scale an Instance Pool size',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an Instance Pool details',
                },
                {
                  name: ['update'],
                  description: 'Update an Instance Pool',
                },
              ],
            },
            {
              name: ['template', 'instance-template'],
              description: 'Compute instance templates management',
              subcommands: [
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Compute instance template',
                },
                {
                  name: ['list'],
                  description: 'List Compute instance templates',
                },
                {
                  name: ['add', 'register'],
                  description:
                    'Register a new Compute instance template',
                },
                {
                  name: ['get', 'show'],
                  description:
                    'Show a Compute instance template details',
                },
              ],
            },
            {
              name: ['instance-type'],
              description: 'Compute instance types management',
              subcommands: [
                {
                  name: ['list'],
                  description: 'List Compute instance types',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a Compute instance type details',
                },
              ],
            },
            {
              name: ['nlb', 'load-balancer'],
              description: 'Network Load Balancers management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create a Network Load Balancer',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Network Load Balancer',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Network Load Balancers',
                },
                {
                  name: ['svc', 'service'],
                  description: 'Manage Network Load Balancer services',
                  subcommands: [
                    {
                      name: ['add'],
                      description:
                        'Add a service to a Network Load Balancer',
                    },
                    {
                      name: ['rm', 'delete'],
                      description:
                        'Delete a Network Load Balancer service',
                    },
                    {
                      name: ['get', 'show'],
                      description:
                        'Show a Network Load Balancer service details',
                    },
                    {
                      name: ['update'],
                      description:
                        'Update a Network Load Balancer service',
                    },
                  ],
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a Network Load Balancer details',
                },
                {
                  name: ['update'],
                  description: 'Update a Network Load Balancer',
                },
              ],
            },
            {
              name: ['privnet', 'private-network'],
              description: 'Private Networks management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create a Private Network',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Private Network',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Private Networks',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a Private Network details',
                },
                {
                  name: ['update'],
                  description: 'Update a Private Network',
                },
              ],
            },
            {
              name: ['sg', 'security-group'],
              description: 'Security Groups management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create a Security Group',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete a Security Group',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List Security Groups',
                },
                {
                  name: ['rule'],
                  description: 'Security Group rules management',
                  subcommands: [
                    {
                      name: ['add'],
                      description: 'Add a Security Group rule',
                    },
                    {
                      name: ['rm', 'delete'],
                      description: 'Delete a Security Group rule',
                    },
                  ],
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a Security Group details',
                },
                {
                  name: ['source'],
                  description:
                    'Security Group external sources management',
                  subcommands: [
                    {
                      name: ['add'],
                      description:
                        'Add an external source to a Security Group',
                    },
                    {
                      name: ['rm', 'remove'],
                      description:
                        'Remove an external source from a Security Group',
                    },
                  ],
                },
              ],
            },
            {
              name: ['sks'],
              description: 'Scalable Kubernetes Service management',
              subcommands: [
                {
                  name: ['authority-cert'],
                  description:
                    'Retrieve an authority certificate for an SKS cluster',
                },
                {
                  name: ['add', 'create'],
                  description: 'Create an SKS cluster',
                },
                {
                  name: ['rm', 'delete'],
                  description: 'Delete an SKS cluster',
                },
                {
                  name: ['dr', 'deprecated-resources'],
                  description:
                    'List resources that will be deprecated in a futur release of Kubernetes for an SKS cluster',
                },
                {
                  name: ['kc', 'kubeconfig'],
                  description:
                    'Generate a Kubernetes kubeconfig file for an SKS cluster',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List SKS clusters',
                },
                {
                  name: ['np', 'nodepool'],
                  description: 'Manage SKS cluster Nodepools',
                  subcommands: [
                    {
                      name: ['add'],
                      description: 'Add a Nodepool to an SKS cluster',
                    },
                    {
                      name: ['rm', 'delete'],
                      description: 'Delete an SKS cluster Nodepool',
                    },
                    {
                      name: ['evict'],
                      description: 'Evict SKS cluster Nodepool members',
                    },
                    {
                      name: ['ls', 'list'],
                      description: 'List SKS cluster Nodepools',
                    },
                    {
                      name: ['scale'],
                      description: 'Scale an SKS cluster Nodepool size',
                    },
                    {
                      name: ['get', 'show'],
                      description:
                        'Show an SKS cluster Nodepool details',
                    },
                    {
                      name: ['update'],
                      description: 'Update an SKS cluster Nodepool',
                    },
                  ],
                },
                {
                  name: ['rotate-ccm-credentials'],
                  description:
                    'Rotate the Exoscale Cloud Controller IAM credentials for an SKS cluster',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an SKS cluster details',
                },
                {
                  name: ['update'],
                  description: 'Update an SKS cluster',
                },
                {
                  name: ['upgrade'],
                  description:
                    'Upgrade an SKS cluster Kubernetes version',
                },
                {
                  name: ['upgrade-service-level'],
                  description: 'Upgrade an SKS cluster service level',
                },
                {
                  name: ['ls', 'versions'],
                  description: 'List supported SKS cluster versions',
                },
              ],
            },
            {
              name: ['ssh-key'],
              description: 'SSH keys management',
              subcommands: [
                {
                  name: ['rm', 'delete'],
                  description: 'Delete an SSH key',
                },
                { name: ['list'], description: 'List SSH keys' },
                {
                  name: ['add', 'register'],
                  description: 'Register an SSH key',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an SSH key details',
                },
              ],
            },
          ],
        },
        {
          name: ['config'],
          description: 'Exoscale CLI configuration management',
          subcommands: [
            {
              name: ['add'],
              description: 'Add a new account to configuration',
            },
            {
              name: ['del', 'delete'],
              description: 'Delete an account from configuration',
            },
            {
              name: ['ls', 'list'],
              description: 'List available accounts',
            },
            {
              name: ['set'],
              description: 'Set an account as default account',
            },
            {
              name: ['get', 'show'],
              description: 'Show an account details',
            },
          ],
        },
        {
          name: ['dbaas'],
          description: 'Database as a Service management',
          subcommands: [
            {
              name: ['ca-certificate'],
              description: 'Retrieve the Database CA certificate',
            },
            {
              name: ['add', 'create'],
              description: 'Create a Database Service',
            },
            {
              name: ['rm', 'delete'],
              description: 'Delete a Database Service',
            },
            {
              name: ['ls', 'list'],
              description: 'List Database Services',
            },
            {
              name: ['get', 'logs'],
              description: 'Query a Database Service logs',
            },
            {
              name: ['get', 'metrics'],
              description: 'Query a Database Service metrics over time',
            },
            {
              name: ['c', 'migration'],
              description: 'migration status/check',
              subcommands: [
                {
                  name: ['status'],
                  description: 'Migration status of a Database',
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a Database Service details',
            },
            {
              name: ['type'],
              description: 'Database Services types management',
              subcommands: [
                {
                  name: ['list'],
                  description: 'List Database Service types',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a Database Service type details',
                },
              ],
            },
            {
              name: ['update'],
              description: 'Update Database Service',
            },
          ],
        },
        {
          name: ['dns'],
          description:
            'DNS cmd lets you host your zones and manage records',
          subcommands: [
            {
              name: ['add'],
              description: 'Add record to domain',
              subcommands: [
                {
                  name: ['A'],
                  description: 'Add A record type to a domain',
                },
                {
                  name: ['AAAA'],
                  description: 'Add AAAA record type to a domain',
                },
                {
                  name: ['ALIAS'],
                  description: 'Add ALIAS record type to a domain',
                },
                {
                  name: ['CAA'],
                  description: 'Add CAA record type to a domain',
                },
                {
                  name: ['CNAME'],
                  description: 'Add CNAME record type to a domain',
                },
                {
                  name: ['HINFO'],
                  description: 'Add HINFO record type to a domain',
                },
                {
                  name: ['MX'],
                  description: 'Add MX record type to a domain',
                },
                {
                  name: ['NAPTR'],
                  description: 'Add NAPTR record type to a domain',
                },
                {
                  name: ['NS'],
                  description: 'Add NS record type to a domain',
                },
                {
                  name: ['POOL'],
                  description: 'Add POOL record type to a domain',
                },
                {
                  name: ['SRV'],
                  description: 'Add SRV record type to a domain',
                },
                {
                  name: ['SSHFP'],
                  description: 'Add SSHFP record type to a domain',
                },
                {
                  name: ['TXT'],
                  description: 'Add TXT record type to a domain',
                },
                {
                  name: ['URL'],
                  description: 'Add URL record type to a domain',
                },
              ],
            },
            { name: ['add', 'create'], description: 'Create a domain' },
            { name: ['del', 'delete'], description: 'Delete a domain' },
            { name: ['ls', 'list'], description: 'List domains' },
            {
              name: ['rm', 'remove'],
              description: 'Remove a domain record',
            },
            { name: ['show'], description: 'Show the domain records' },
            {
              name: ['update'],
              description: 'Update domain record',
              subcommands: [
                {
                  name: ['A'],
                  description: 'Update A record type to a domain',
                },
                {
                  name: ['AAAA'],
                  description: 'Update AAAA record type to a domain',
                },
                {
                  name: ['ALIAS'],
                  description: 'Update ALIAS record type to a domain',
                },
                {
                  name: ['CNAME'],
                  description: 'Update CNAME record type to a domain',
                },
                {
                  name: ['HINFO'],
                  description: 'Update HINFO record type to a domain',
                },
                {
                  name: ['MX'],
                  description: 'Update MX record type to a domain',
                },
                {
                  name: ['NAPTR'],
                  description: 'Update NAPTR record type to a domain',
                },
                {
                  name: ['NS'],
                  description: 'Update NS record type to a domain',
                },
                {
                  name: ['POOL'],
                  description: 'Update POOL record type to a domain',
                },
                {
                  name: ['SPF'],
                  description: 'Update SPF record type to a domain',
                },
                {
                  name: ['SRV'],
                  description: 'Update SRV record type to a domain',
                },
                {
                  name: ['SSHFP'],
                  description: 'Update SSHFP record type to a domain',
                },
                {
                  name: ['TXT'],
                  description: 'Update TXT record type to a domain',
                },
                {
                  name: ['URL'],
                  description: 'Update URL record type to a domain',
                },
              ],
            },
          ],
        },
        {
          name: ['environment'],
          description: 'Environment variables usage',
        },
        {
          name: ['iam'],
          description: 'Identity and Access Management',
          subcommands: [
            {
              name: ['key', 'access-key'],
              description: 'IAM access keys management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create an IAM access key',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List IAM access keys',
                },
                {
                  name: ['ls', 'list-operations'],
                  description: 'List IAM access keys operations',
                },
                {
                  name: ['add', 'revoke'],
                  description: 'Revoke an IAM access key',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an IAM access key details',
                },
              ],
            },
          ],
        },
        {
          name: ['lab'],
          description: 'Experimental commands',
          subcommands: [
            {
              name: ['coi'],
              description: 'Deploy a Container-Optimized Instance',
            },
          ],
        },
        { name: ['limits'], description: 'Current account limits' },
        { name: ['output'], description: 'Output formatting usage' },
        {
          name: ['runstatus'],
          description: 'Manage your Runstat.us pages',
          subcommands: [
            {
              name: ['add', 'create'],
              description: 'Create Runstat.us page',
            },
            {
              name: ['del', 'delete'],
              description: 'Delete runstat.us page(s)',
            },
            {
              name: ['incident'],
              description: 'Incident management',
              subcommands: [
                {
                  name: ['add'],
                  description: 'Add an incident to a runstat.us page',
                },
                { name: ['ls', 'list'], description: 'List incidents' },
                {
                  name: ['rm', 'remove'],
                  description: 'Remove incident from a runstat.us page',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show an incident details',
                },
                {
                  name: ['update'],
                  description: 'update an existing incident',
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List runstat.us pages',
            },
            {
              name: ['maintenance'],
              description: 'Maintenance management',
              subcommands: [
                {
                  name: ['add'],
                  description: 'Add a maintenance to a runstat.us page',
                },
                {
                  name: ['ls', 'list'],
                  description: 'List maintenance from page(s)',
                },
                {
                  name: ['rm', 'remove'],
                  description:
                    'Remove maintenance from a runstat.us page',
                },
                {
                  name: ['get', 'show'],
                  description: 'Show a maintenance details',
                },
                {
                  name: ['update'],
                  description: 'update a maintenance',
                },
              ],
            },
            {
              name: ['service'],
              description: 'Runstat.us service management',
              subcommands: [
                {
                  name: ['add', 'create'],
                  description: 'Create a service',
                },
                {
                  name: ['del', 'delete'],
                  description: 'Delete a service',
                },
                { name: ['ls', 'list'], description: 'List services' },
                {
                  name: ['get', 'show'],
                  description: 'Show a service details',
                },
              ],
            },
            {
              name: ['get', 'show'],
              description: 'Show a runstat.us page details',
            },
          ],
        },
        { name: ['status'], description: 'Exoscale status' },
        {
          name: ['storage'],
          description: 'Object Storage management',
          subcommands: [
            {
              name: ['cors'],
              description: 'Manage buckets CORS configuration',
              subcommands: [
                {
                  name: ['add'],
                  description:
                    'Add a CORS configuration rule to a bucket',
                },
                {
                  name: ['del', 'delete'],
                  description:
                    'Delete the CORS configuration of a bucket',
                },
              ],
            },
            {
              name: ['del', 'rm', 'delete'],
              description: 'Delete objects',
            },
            {
              name: ['get', 'download'],
              description: 'Download files from a bucket',
            },
            {
              name: ['headers'],
              description: 'Manage objects HTTP headers',
              subcommands: [
                {
                  name: ['add'],
                  description: 'Add HTTP headers to an object',
                },
                {
                  name: ['del', 'delete'],
                  description: 'Delete HTTP headers from an object',
                },
              ],
            },
            {
              name: ['ls', 'list'],
              description: 'List buckets and objects',
            },
            {
              name: ['create', 'mb'],
              description: 'Create a new bucket',
            },
            {
              name: ['meta', 'metadata'],
              description: 'Manage objects metadata',
              subcommands: [
                {
                  name: ['add'],
                  description: 'Add key/value metadata to an object',
                },
                {
                  name: ['del', 'delete'],
                  description: 'Delete metadata from an object',
                },
              ],
            },
            {
              name: ['presign'],
              description: 'Generate a pre-signed URL to an object',
            },
            { name: ['rb'], description: 'Delete a bucket' },
            {
              name: ['setacl'],
              description: 'Set a bucket/objects ACL',
            },
            {
              name: ['show'],
              description: 'Show a bucket/object details',
            },
            {
              name: ['put', 'upload'],
              description: 'Upload files to a bucket',
            },
          ],
        },
        { name: ['version'], description: 'Print the version of exo' },
        {
          name: ['zones', 'zone'],
          description: 'List all available zones',
        },
      ],
    },
  ],
  options: [
    {
      name: ['--config', '-C'],
      description:
        'Specify an alternate config file [env EXOSCALE_CONFIG]',
      isPersistent: true,
      args: [{ name: 'config' }],
    },
    {
      name: ['--output-format', '-O'],
      description:
        'Output format (table|json|text), see "exo output --help" for more information',
      isPersistent: true,
      args: [{ name: 'output-format' }],
    },
    {
      name: ['--output-template'],
      description: 'Template to use if output format is "text"',
      isPersistent: true,
      args: [{ name: 'output-template' }],
    },
    {
      name: ['--quiet', '-Q'],
      description: 'Quiet mode (disable non-essential command output)',
      isPersistent: true,
    },
    {
      name: ['--use-account', '-A'],
      description:
        'Account to use in config file [env EXOSCALE_ACCOUNT]',
      isPersistent: true,
      args: [{ name: 'use-account' }],
    },
    {
      name: ['--help', '-h'],
      description: 'Display help',
      isPersistent: true,
    },
  ],
};
export default completionSpec;
