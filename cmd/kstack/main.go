package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	cfg "github.com/christk1/kstack/internal/config"
	"github.com/christk1/kstack/pkg/addons"
	_ "github.com/christk1/kstack/pkg/addons/exampleapp"
	_ "github.com/christk1/kstack/pkg/addons/grafana"
	_ "github.com/christk1/kstack/pkg/addons/kafka"
	pgaddon "github.com/christk1/kstack/pkg/addons/postgres"
	_ "github.com/christk1/kstack/pkg/addons/prometheus"
	"github.com/christk1/kstack/pkg/cluster"
	"github.com/christk1/kstack/pkg/helm"
	"github.com/christk1/kstack/pkg/preflight"
	"github.com/christk1/kstack/utils"
)

// Version info populated via -ldflags at build time
var (
	version   = "dev"
	commit    = ""
	buildDate = ""
)

type rootOptions struct {
	provider    string
	clusterName string
	addons      string
	namespace   string
	kubeconfig  string
	helmPath    string
	timeout     time.Duration
	verbose     bool
	debug       bool
	noColor     bool
	dryRun      bool
}

func main() {
	opts := &rootOptions{}

	rootCmd := &cobra.Command{
		Use:           "kstack",
		Short:         "Create local K8s clusters and install Helm-based addons",
		Long:          "kstack is a developer-first CLI to spin up local kind/k3d clusters and install addons like Prometheus, Kafka, and Postgres via Helm.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&opts.provider, "provider", "kind", "Cluster provider: kind|k3d")
	rootCmd.PersistentFlags().StringVar(&opts.clusterName, "cluster", "kstack", "Cluster name")
	rootCmd.PersistentFlags().StringVar(&opts.addons, "addons", "", "Comma-separated addons to install (e.g. prometheus,postgres)")
	rootCmd.PersistentFlags().StringVar(&opts.namespace, "namespace", "kstack", "Kubernetes namespace for addons")
	rootCmd.PersistentFlags().StringVar(&opts.kubeconfig, "kubeconfig", "", "Path to kubeconfig (optional)")
	rootCmd.PersistentFlags().StringVar(&opts.helmPath, "helm", "helm", "Path to helm binary")
	rootCmd.PersistentFlags().DurationVar(&opts.timeout, "timeout", 10*time.Minute, "Overall operation timeout")
	rootCmd.PersistentFlags().BoolVarP(&opts.verbose, "verbose", "v", false, "Verbose logging")
	rootCmd.PersistentFlags().BoolVar(&opts.debug, "debug", false, "Print resolved configuration and extra diagnostics")
	rootCmd.PersistentFlags().BoolVar(&opts.noColor, "no-color", false, "Disable ANSI colors in logs")
	rootCmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", false, "Print planned actions without executing external commands")

	// Subcommands
	rootCmd.AddCommand(newUpCmd(opts))
	rootCmd.AddCommand(newDownCmd(opts))
	rootCmd.AddCommand(newAddonsCmd(opts))
	rootCmd.AddCommand(newPreflightCmd(opts))
	rootCmd.AddCommand(newStatusCmd(opts))
	rootCmd.AddCommand(newVersionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newUpCmd(opts *rootOptions) *cobra.Command {
	var setPairs []string
	var extraValues []string
	var ha bool

	cmd := &cobra.Command{
		Use:   "up",
		Short: "Create cluster and install addons",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Build configuration from flags and env
			c := cfg.Defaults()
			c.Provider = opts.provider
			c.ClusterName = opts.clusterName
			c.Addons = cfg.ParseAddonsCSV(opts.addons)
			c.Namespace = opts.namespace
			c.Kubeconfig = opts.kubeconfig
			c.HelmPath = opts.helmPath
			c.Timeout = opts.timeout
			c.Verbose = opts.verbose
			c.Debug = opts.debug
			c = cfg.FromEnv(c)

			utils.SetVerbose(c.Verbose)
			utils.SetColorEnabled(!opts.noColor)
			if opts.dryRun {
				utils.Info("DRY-RUN: no external commands will be executed")
			}
			utils.Debug("config: provider=%s cluster=%s ns=%s addons=%v kubeconfig=%s helm=%s timeout=%s verbose=%v", c.Provider, c.ClusterName, c.Namespace, c.Addons, c.Kubeconfig, c.HelmPath, c.Timeout, c.Verbose)

			// Instantiate provider
			var prov cluster.Provider
			switch c.Provider {
			case "kind":
				prov = cluster.NewKindProvider(c.ClusterName)
			case "k3d":
				prov = cluster.NewK3dProvider(c.ClusterName)
			default:
				return fmt.Errorf("unknown provider: %s", c.Provider)
			}

			// apply dry-run to providers if supported
			cluster.SetDryRun(prov, opts.dryRun)

			ctx, cancel := context.WithTimeout(cmd.Context(), c.Timeout)
			defer cancel()

			// preflight checks
			preflight.Verbose = c.Verbose
			preflight.Debug = c.Debug

			if !opts.dryRun {
				if err := preflight.CheckProviderCLI(c.Provider); err != nil {
					return err
				}
				if err := preflight.CheckDocker(ctx); err != nil {
					return err
				}
			} else {
				utils.Info("DRY-RUN: check provider CLI for %s", c.Provider)
				utils.Info("DRY-RUN: check docker info")
			}

			exists, err := prov.Exists(ctx)
			if err != nil {
				return fmt.Errorf("failed to query existing cluster: %w", err)
			}
			if !exists {
				utils.Info("creating cluster %s with provider %s", c.ClusterName, c.Provider)
				var sp *utils.Spinner
				if !opts.dryRun {
					sp = utils.NewSpinner("Creating cluster")
					sp.Start()
				}
				if err := prov.Create(ctx); err != nil {
					if sp != nil {
						sp.Stop()
					}
					return fmt.Errorf("create failed: %w", err)
				}
				if sp != nil {
					sp.Stop()
				}
			} else {
				utils.Info("cluster %s already exists", c.ClusterName)
			}

			kubePath, err := prov.KubeconfigPath(ctx)
			if err == nil && kubePath != "" {
				utils.Info("kubeconfig available at: %s", kubePath)
			} else if err != nil {
				utils.Debug("kubeconfig not available: %v", err)
			}

			if len(c.Addons) > 0 {
				hc := helm.NewClient(c.HelmPath)
				hc.DryRun = opts.dryRun
				if ver, err := hc.Preflight(10 * time.Second); err != nil {
					return fmt.Errorf("helm not available or fails preflight: %w", err)
				} else {
					utils.Debug("helm version: %s", ver)
				}

				// validate user-supplied values files first
				for _, vf := range extraValues {
					if vf == "" {
						continue
					}
					info, err := os.Stat(vf)
					if err != nil {
						return fmt.Errorf("values file %s not found or unreadable: %w", vf, err)
					}
					if !info.Mode().IsRegular() {
						return fmt.Errorf("values file %s is not a regular file", vf)
					}
				}

				// validate --set pairs
				var badSets []string
				for _, s := range setPairs {
					if s == "" || !strings.Contains(s, "=") {
						badSets = append(badSets, s)
						continue
					}
					parts := strings.SplitN(s, "=", 2)
					if strings.TrimSpace(parts[0]) == "" {
						badSets = append(badSets, s)
					}
				}
				if len(badSets) > 0 {
					return fmt.Errorf("invalid --set entries: %v; expected key=val with non-empty key", badSets)
				}

				for _, name := range c.Addons {
					utils.Info("installing addon %s...", name)
					a, err := addons.Get(name)
					if err != nil {
						return err
					}
					var sp *utils.Spinner
					if !opts.dryRun {
						sp = utils.NewSpinner(fmt.Sprintf("Installing %s", a.Name()))
						sp.Start()
					}

					chartName := a.Chart()
					repoName := a.RepoName()
					repoURL := a.RepoURL()
					if ha && a.Name() == "postgres" {
						chartName = "bitnami/postgresql-ha"
						repoName = "bitnami"
						repoURL = "https://charts.bitnami.com/bitnami"
					}

					if !(strings.HasPrefix(chartName, "./") || strings.HasPrefix(chartName, "/")) {
						if err := hc.RepoAdd(repoName, repoURL); err != nil {
							if sp != nil {
								sp.Stop()
							}
							return err
						}
						if err := hc.RepoUpdate(); err != nil {
							if sp != nil {
								sp.Stop()
							}
							return err
						}
					}
					vals := a.ValuesFiles()
					if ha && a.Name() == "postgres" {
						if haFile, err := pgaddon.HAValuesFile(); err == nil && haFile != "" {
							vals = append(vals, haFile)
						} else if err != nil {
							utils.Debug("failed to load HA values for postgres: %v", err)
						}
					}
					allValues := append(vals, extraValues...)
					merged, cleanup, err := helm.MergeValues(allValues, nil)
					if err != nil {
						return err
					}
					if err := hc.InstallOrUpgrade(a.Name(), chartName, a.Namespace(), merged, true, 15*time.Minute, false, setPairs); err != nil {
						if cerr := cleanup(); cerr != nil {
							utils.Debug("cleanup error: %v", cerr)
						}
						if sp != nil {
							sp.Stop()
						}
						return err
					}
					if cerr := cleanup(); cerr != nil {
						utils.Debug("cleanup error: %v", cerr)
					}
					if sp != nil {
						sp.Stop()
					}
					utils.Info("addon %s installed", name)
				}
			} else {
				utils.Info("no addons requested")
			}

			return nil
		},
	}
	cmd.Flags().StringArrayVar(&setPairs, "set", nil, "Set values (key=val). Can be supplied multiple times")
	cmd.Flags().StringArrayVar(&extraValues, "values", nil, "Additional values files to pass (-f) to Helm. Can be supplied multiple times")
	cmd.Flags().BoolVar(&ha, "ha", false, "Install HA variant for supported addons (e.g. postgres)")
	return cmd
}

func newDownCmd(opts *rootOptions) *cobra.Command {
	var purgeAddons bool
	cmd := &cobra.Command{
		Use:   "down",
		Short: "Delete cluster (and optionally uninstall addons)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cfg.Defaults()
			c.Provider = opts.provider
			c.ClusterName = opts.clusterName
			c.Namespace = opts.namespace
			c.Timeout = opts.timeout
			c.Verbose = opts.verbose
			c.Debug = opts.debug
			c = cfg.FromEnv(c)

			utils.SetVerbose(c.Verbose)
			utils.SetColorEnabled(!opts.noColor)
			if opts.dryRun {
				utils.Info("DRY-RUN: no external commands will be executed")
			}

			var prov cluster.Provider
			switch c.Provider {
			case "kind":
				prov = cluster.NewKindProvider(c.ClusterName)
			case "k3d":
				prov = cluster.NewK3dProvider(c.ClusterName)
			default:
				return fmt.Errorf("unknown provider: %s", c.Provider)
			}
			cluster.SetDryRun(prov, opts.dryRun)

			ctx, cancel := context.WithTimeout(cmd.Context(), c.Timeout)
			defer cancel()

			if purgeAddons {
				utils.Info("purging addons before cluster deletion")
				hc := helm.NewClient(opts.helmPath)
				hc.DryRun = opts.dryRun
				if ver, err := hc.Preflight(10 * time.Second); err != nil {
					return fmt.Errorf("helm is required to purge addons: %w", err)
				} else {
					utils.Debug("helm version: %s", ver)
				}
				for _, name := range addons.List() {
					a, err := addons.Get(name)
					if err != nil {
						utils.Debug("skipping uninstall for %s: %v", name, err)
						continue
					}
					if err := hc.Uninstall(a.Name(), a.Namespace(), true, 30*time.Second); err != nil {
						utils.Debug("failed to uninstall addon %s: %v", name, err)
						continue
					}
					utils.Info("uninstalled addon %s", name)
				}
			}

			utils.Info("deleting cluster %s (purgeAddons=%v)", c.ClusterName, purgeAddons)
			var sp *utils.Spinner
			if !opts.dryRun {
				sp = utils.NewSpinner("Deleting cluster")
				sp.Start()
			}
			if err := prov.Delete(ctx); err != nil {
				if sp != nil {
					sp.Stop()
				}
				return fmt.Errorf("delete failed: %w", err)
			}
			if sp != nil {
				sp.Stop()
			}
			utils.Info("cluster %s deleted", c.ClusterName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&purgeAddons, "purge-addons", false, "Uninstall addons before cluster deletion")
	return cmd
}

func newAddonsCmd(opts *rootOptions) *cobra.Command {
	addonsCmd := &cobra.Command{Use: "addons", Short: "Manage addons (install/uninstall/list)"}

	var waitForInstall bool
	var helmTimeout time.Duration
	var atomicInstall bool
	var setPairs []string
	var extraValues []string

	var installHA bool
	installCmd := &cobra.Command{
		Use:   "install [name]",
		Short: "Install an addon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			utils.SetVerbose(opts.verbose)
			utils.SetColorEnabled(!opts.noColor)
			if opts.dryRun {
				utils.Info("DRY-RUN: no external commands will be executed")
			}

			hc := helm.NewClient(opts.helmPath)
			hc.DryRun = opts.dryRun
			if ver, err := hc.Preflight(10 * time.Second); err != nil {
				return fmt.Errorf("helm not available or fails preflight: %w", err)
			} else {
				utils.Debug("helm version: %s", ver)
			}

			a, err := addons.Get(name)
			if err != nil {
				return err
			}

			chartName := a.Chart()
			repoName := a.RepoName()
			repoURL := a.RepoURL()
			if installHA && a.Name() == "postgres" {
				chartName = "bitnami/postgresql-ha"
				repoName = "bitnami"
				repoURL = "https://charts.bitnami.com/bitnami"
			}

			if !(strings.HasPrefix(chartName, "./") || strings.HasPrefix(chartName, "/")) {
				if err := hc.RepoAdd(repoName, repoURL); err != nil {
					return err
				}
				if err := hc.RepoUpdate(); err != nil {
					return err
				}
			}

			for _, vf := range extraValues {
				if vf == "" {
					continue
				}
				info, err := os.Stat(vf)
				if err != nil {
					return fmt.Errorf("values file %s not found or unreadable: %w", vf, err)
				}
				if !info.Mode().IsRegular() {
					return fmt.Errorf("values file %s is not a regular file", vf)
				}
			}

			var badSets []string
			for _, s := range setPairs {
				if s == "" || !strings.Contains(s, "=") {
					badSets = append(badSets, s)
					continue
				}
				parts := strings.SplitN(s, "=", 2)
				if strings.TrimSpace(parts[0]) == "" {
					badSets = append(badSets, s)
				}
			}
			if len(badSets) > 0 {
				return fmt.Errorf("invalid --set entries: %v; expected key=val with non-empty key", badSets)
			}

			addonValues := a.ValuesFiles()
			if installHA && a.Name() == "postgres" {
				if haFile, err := pgaddon.HAValuesFile(); err == nil && haFile != "" {
					addonValues = append(addonValues, haFile)
				} else if err != nil {
					utils.Debug("failed to load HA values for postgres: %v", err)
				}
			}
			allValues := append(addonValues, extraValues...)
			merged, cleanup, err := helm.MergeValues(allValues, nil)
			if err != nil {
				return err
			}

			release := a.Name()
			if err := hc.InstallOrUpgrade(release, chartName, a.Namespace(), merged, waitForInstall, helmTimeout, atomicInstall, setPairs); err != nil {
				if cerr := cleanup(); cerr != nil {
					utils.Debug("cleanup after failed install had error: %v", cerr)
				}
				return err
			}
			utils.Info("installed addon %s (release=%s) in ns=%s", a.Name(), release, a.Namespace())
			if cerr := cleanup(); cerr != nil {
				utils.Debug("cleanup after install had error: %v", cerr)
			}
			return nil
		},
	}
	installCmd.Flags().BoolVar(&waitForInstall, "wait", false, "Wait for resources to become ready (passes --wait to helm)")
	installCmd.Flags().DurationVar(&helmTimeout, "helm-timeout", 15*time.Minute, "Timeout passed to helm --timeout when --wait is set")
	installCmd.Flags().BoolVar(&atomicInstall, "atomic", false, "Use --atomic with helm upgrade --install")
	installCmd.Flags().StringArrayVar(&setPairs, "set", nil, "Set values (key=val). Can be supplied multiple times")
	installCmd.Flags().StringArrayVar(&extraValues, "values", nil, "Additional values files to pass (-f) to Helm. Can be supplied multiple times")
	installCmd.Flags().BoolVar(&installHA, "ha", false, "Install HA variant for supported addons (e.g. postgres)")

	var uninstallWait bool
	var uninstallTimeout time.Duration
	uninstallCmd := &cobra.Command{
		Use:   "uninstall [name]",
		Short: "Uninstall an addon",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			utils.SetVerbose(opts.verbose)
			utils.SetColorEnabled(!opts.noColor)
			if opts.dryRun {
				utils.Info("DRY-RUN: no external commands will be executed")
			}
			hc := helm.NewClient(opts.helmPath)
			hc.DryRun = opts.dryRun
			if ver, err := hc.Preflight(10 * time.Second); err != nil {
				return fmt.Errorf("helm not available or fails preflight: %w", err)
			} else {
				utils.Debug("helm version: %s", ver)
			}
			a, err := addons.Get(name)
			if err != nil {
				return err
			}
			if err := hc.Uninstall(a.Name(), a.Namespace(), uninstallWait, uninstallTimeout); err != nil {
				return err
			}
			utils.Info("uninstalled addon %s from ns=%s", a.Name(), a.Namespace())
			return nil
		},
	}
	uninstallCmd.Flags().BoolVar(&uninstallWait, "wait", false, "Wait for uninstall to complete (passes --timeout to helm)")
	uninstallCmd.Flags().DurationVar(&uninstallTimeout, "helm-timeout", 30*time.Second, "Timeout passed to helm for uninstall when --wait is set")

	listCmd := &cobra.Command{Use: "list", Short: "List available addons", RunE: func(cmd *cobra.Command, args []string) error {
		utils.SetVerbose(opts.verbose)
		utils.SetColorEnabled(!opts.noColor)
		names := addons.List()
		utils.Info("Available addons: %v", names)
		return nil
	}}

	addonsCmd.AddCommand(installCmd, uninstallCmd, listCmd)
	return addonsCmd
}

func newStatusCmd(opts *rootOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show cluster and addon status",
		RunE: func(cmd *cobra.Command, args []string) error {
			utils.SetVerbose(opts.verbose)
			utils.SetColorEnabled(!opts.noColor)
			if opts.dryRun {
				utils.Info("DRY-RUN: no external commands will be executed")
			}

			c := cfg.Defaults()
			c.Provider = opts.provider
			c.ClusterName = opts.clusterName
			c.Namespace = opts.namespace
			c.Kubeconfig = opts.kubeconfig
			c.HelmPath = opts.helmPath
			c.Timeout = opts.timeout
			c.Verbose = opts.verbose
			c.Debug = opts.debug
			c = cfg.FromEnv(c)

			utils.Info("status for %s cluster '%s' (ns=%s)", c.Provider, c.ClusterName, c.Namespace)

			var prov cluster.Provider
			switch c.Provider {
			case "kind":
				prov = cluster.NewKindProvider(c.ClusterName)
			case "k3d":
				prov = cluster.NewK3dProvider(c.ClusterName)
			default:
				return fmt.Errorf("unknown provider: %s", c.Provider)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), c.Timeout)
			defer cancel()
			exists, err := prov.Exists(ctx)
			if err != nil {
				utils.Info("- cluster: error checking existence: %v", err)
			} else if exists {
				utils.Info("- cluster: exists")
				if kp, err := prov.KubeconfigPath(ctx); err == nil && kp != "" {
					utils.Info("  kubeconfig: %s", kp)
				}
			} else {
				utils.Info("- cluster: not found")
			}

			hc := helm.NewClient(c.HelmPath)
			hc.DryRun = opts.dryRun
			if ver, err := hc.Preflight(5 * time.Second); err != nil {
				utils.Info("- helm: not available (%v)", err)
			} else {
				utils.Debug("helm version: %s", ver)
				rels, err := hc.ListReleases(c.Namespace)
				if err != nil {
					utils.Info("- addons: failed to list releases: %v", err)
				} else if len(rels) == 0 {
					utils.Info("- addons: none (namespace=%s)", c.Namespace)
				} else {
					utils.Info("- addons (namespace=%s):", c.Namespace)
					for _, r := range rels {
						utils.Info("  - %s: %s (chart=%s) updated=%s", r.Name, r.Status, r.Chart, r.Updated)
					}
				}
			}
			return nil
		},
	}
	return cmd
}

func newPreflightCmd(opts *rootOptions) *cobra.Command {
	var verboseFlag bool
	var debugFlag bool
	cmd := &cobra.Command{
		Use:   "preflight",
		Short: "Run environment preflight checks (CLIs, Docker, Helm, provider)",
		RunE: func(cmd *cobra.Command, args []string) error {
			c := cfg.Defaults()
			c.Provider = opts.provider
			c.HelmPath = opts.helmPath
			c.Timeout = opts.timeout
			c.Verbose = opts.verbose
			c.Debug = opts.debug
			c = cfg.FromEnv(c)

			utils.SetVerbose(opts.verbose)
			utils.SetColorEnabled(!opts.noColor)
			if opts.dryRun {
				utils.Info("DRY-RUN: no external commands will be executed")
			}

			preflight.Verbose = verboseFlag || c.Verbose
			preflight.Debug = debugFlag || c.Debug

			if !opts.dryRun {
				if err := preflight.CheckProviderCLI(c.Provider); err != nil {
					return err
				}
			} else {
				utils.Info("DRY-RUN: preflight provider CLI check for %s", c.Provider)
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()
			if !opts.dryRun {
				if err := preflight.CheckDocker(ctx); err != nil {
					return err
				}
			} else {
				utils.Info("DRY-RUN: preflight docker check")
			}

			hc := helm.NewClient(c.HelmPath)
			hc.DryRun = opts.dryRun
			if ver, err := hc.Preflight(5 * time.Second); err != nil {
				return fmt.Errorf("helm preflight failed: %w", err)
			} else {
				fmt.Printf("helm: %s\n", ver)
			}

			fmt.Println("preflight checks passed")
			return nil
		},
	}
	cmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Print full command output from checks")
	cmd.Flags().BoolVar(&debugFlag, "debug", false, "Print debug output from checks")
	return cmd
}

func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "version", Short: "Print version information", RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("kstack %s", version)
		if commit != "" {
			fmt.Printf(" (commit %s)", commit)
		}
		if buildDate != "" {
			fmt.Printf(" built %s", buildDate)
		}
		fmt.Println()
		return nil
	}}
	return cmd
}
