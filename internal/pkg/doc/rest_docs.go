//Copyright 2015 Red Hat Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func GenReSTTree(cmd *cobra.Command, dir string, linkHandler func(string) string, depth int) error {
	path := filepath.Join(strings.Split(cmd.CommandPath(), " ")...)

	if cmd.HasSubCommands() {
		file := filepath.Join(dir, path, "index.rst")

		if err := os.MkdirAll(filepath.Dir(file), 0777); err != nil {
			return err
		}

		if err := GenReSTIndex(cmd, file, indexHeader, SphinxRef); err != nil {
			return err
		}

		for _, c := range cmd.Commands() {
			if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
				continue
			}

			if err := GenReSTTree(c, dir, linkHandler, depth+1); err != nil {
				return err
			}
		}

		return nil
	}

	name := strings.ReplaceAll(cmd.CommandPath(), " ", "_") + ".rst"
	file := filepath.Join(dir, filepath.Dir(path), name)

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return GenReST(cmd, f, linkHandler, depth)
}

func GenReST(cmd *cobra.Command, w io.Writer, linkHandler func(string) string, depth int) error {
	cmd.InitDefaultHelpCmd()
	cmd.InitDefaultHelpFlag()

	buf := new(bytes.Buffer)

	name := cmd.CommandPath()
	ref := strings.ReplaceAll(name, " ", "_")

	buf.WriteString(fmt.Sprintf(".. _%s:\n\n", ref))
	buf.WriteString(name + "\n")
	buf.WriteString(strings.Repeat("-", len(name)) + "\n\n")

	printWarnings(buf, cmd, depth)

	desc := cmd.Short
	if cmd.Long != "" {
		desc = cmd.Long
	}
	buf.WriteString("Description\n")
	buf.WriteString("~~~~~~~~~~~\n\n")
	buf.WriteString(desc + "\n\n")

	if cmd.Runnable() {
		buf.WriteString(fmt.Sprintf("::\n\n  %s\n\n", cmd.UseLine()))
	}

	printTips(buf, cmd, depth)

	if err := printOptions(buf, cmd); err != nil {
		return err
	}

	if hasSeeAlso(cmd) {
		buf.WriteString("See Also\n")
		buf.WriteString("~~~~~~~~\n\n")
		if cmd.HasParent() {
			parent := cmd.Parent()

			ref = strings.ReplaceAll(parent.CommandPath(), " ", "_")
			if cmd.Root() == parent {
				ref += "-ref"
			}

			buf.WriteString(fmt.Sprintf("* %s - %s\n", linkHandler(ref), parent.Short))

			cmd.VisitParents(func(c *cobra.Command) {
				if c.DisableAutoGenTag {
					cmd.DisableAutoGenTag = c.DisableAutoGenTag
				}
			})
		}

		children := cmd.Commands()
		sort.Sort(byName(children))

		for _, child := range children {
			if !child.IsAvailableCommand() || child.IsAdditionalHelpTopicCommand() {
				continue
			}
			cname := name + " " + child.Name()
			ref = strings.ReplaceAll(cname, " ", "_")
			buf.WriteString(fmt.Sprintf("* %s - %s\n", linkHandler(ref), child.Short))
		}
		buf.WriteString("\n")
	}
	if !cmd.DisableAutoGenTag {
		buf.WriteString("*Auto generated by spf13/cobra on " + time.Now().Format("2-Jan-2006") + "*\n")
	}
	_, err := buf.WriteTo(w)
	return err
}

func printOptions(buf *bytes.Buffer, cmd *cobra.Command) error {
	pcmd.LabelRequiredFlags(cmd)

	flags := cmd.NonInheritedFlags()
	flags.SetOutput(buf)
	if flags.HasAvailableFlags() {
		buf.WriteString("Flags\n")
		buf.WriteString("~~~~~\n\n")
		buf.WriteString("::\n\n")
		flags.PrintDefaults()
		buf.WriteString("\n")
	}

	parentFlags := cmd.InheritedFlags()
	parentFlags.SetOutput(buf)
	if parentFlags.HasAvailableFlags() {
		buf.WriteString("Global Flags\n")
		buf.WriteString("~~~~~~~~~~~~\n\n")
		buf.WriteString("::\n\n")
		parentFlags.PrintDefaults()
		buf.WriteString("\n")
	}

	if len(cmd.Example) > 0 {
		buf.WriteString("Examples\n")
		buf.WriteString("~~~~~~~~\n\n")
		buf.WriteString(cmd.Example)
	}

	return nil
}

func printWarnings(buf *bytes.Buffer, cmd *cobra.Command, depth int) {
	if strings.HasPrefix(cmd.CommandPath(), "confluent local") {
		include := strings.Repeat("../", depth) + "includes/cli.rst"
		args := map[string]string{
			"start-after": "cli_limitations_start",
			"end-before":  "cli_limitations_end",
		}
		buf.WriteString(sphinxBlock("include", include, args))
	}
}

func printTips(buf *bytes.Buffer, cmd *cobra.Command, depth int) {
	if strings.HasPrefix(cmd.CommandPath(), "confluent local") {
		include := strings.Repeat("../", depth) + "includes/path-set-cli.rst"
		buf.WriteString(sphinxBlock("include", include, nil))
	}

	if strings.HasPrefix(cmd.CommandPath(), "confluent secret") {
		ref := SphinxRef("secrets-examples")
		tip := fmt.Sprintf("For examples, see %s.", ref)
		buf.WriteString(sphinxBlock("tip", tip, nil))
	}

	if cmd.CommandPath() == "confluent iam rolebinding create" {
		ref := SphinxRef("view-audit-logs-on-the-fly")
		note := fmt.Sprintf("If you need to troubleshoot when setting up role bindings, it may be helpful to view audit logs on the fly to identify authorization events for specific principals, resources, or operations. For details, refer to %s.", ref)
		buf.WriteString(sphinxBlock("note", note, nil))
	}
}

func SphinxRef(ref string) string {
	return fmt.Sprintf(":ref:`%s`", ref)
}

func sphinxBlock(key, val string, args map[string]string) string {
	str := strings.Builder{}

	str.WriteString(fmt.Sprintf(".. %s:: %s\n", key, val))

	var keys []string
	for key, _ := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		val := args[key]
		str.WriteString(fmt.Sprintf("  :%s: %s\n", key, val))
	}

	str.WriteString("\n")

	return str.String()
}
