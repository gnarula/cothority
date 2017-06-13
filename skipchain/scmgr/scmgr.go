/*
* The skipchain-manager lets you create, modify and query skipchains
 */
package main

import (
	"os"

	"github.com/dedis/cothority/skipchain"

	"gopkg.in/dedis/onet.v1/app"

	"fmt"
	"io/ioutil"

	"errors"

	"encoding/hex"

	"path"

	"bytes"
	"sort"

	"strings"

	"encoding/base64"

	"gopkg.in/dedis/onet.v1/log"
	"gopkg.in/dedis/onet.v1/network"
	"gopkg.in/urfave/cli.v1"
)

type config struct {
	Sbb *skipchain.SBBStorage
}

func main() {
	network.RegisterMessage(&config{})
	cliApp := cli.NewApp()
	cliApp.Name = "scmgr"
	cliApp.Usage = "Create, modify and query skipchains"
	cliApp.Version = "0.1"
	groupsDef := "the group-definition-file"
	cliApp.Commands = []cli.Command{
		{
			Name:      "create",
			Usage:     "make a new skipchain",
			Aliases:   []string{"c"},
			ArgsUsage: groupsDef,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "base, b",
					Value: 2,
					Usage: "base for skipchains",
				},
				cli.IntFlag{
					Name:  "height, he",
					Value: 2,
					Usage: "maximum height of skipchain",
				},
				cli.StringFlag{
					Name:  "url",
					Usage: "URL of html-skipchain",
				},
			},
			Action: create,
		},
		{
			Name:      "join",
			Usage:     "join a skipchain and store it locally",
			Aliases:   []string{"j"},
			ArgsUsage: groupsDef + " skipchain-id",
			Action:    join,
		},
		{
			Name:      "add",
			Usage:     "add a new roster to a skipchain",
			Aliases:   []string{"a"},
			ArgsUsage: "skipchain-id " + groupsDef,
			Action:    add,
		},
		{
			Name:      "update",
			Usage:     "get latest valid block",
			Aliases:   []string{"u"},
			ArgsUsage: "skipchain-id",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "data, d",
				},
			},
			Action: update,
		},
		{
			Name:  "list",
			Usage: "handle list of skipblocks",
			Subcommands: []cli.Command{
				{
					Name:    "known",
					Aliases: []string{"k"},
					Usage:   "lists all known skipblocks",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "long, l",
							Usage: "give long id of blocks",
						},
					},
					Action: lsKnown,
				},
				{
					Name:      "fetch",
					Usage:     "ask all known conodes for skipchains",
					ArgsUsage: "[group-file]",
					Flags: []cli.Flag{
						cli.BoolFlag{
							Name:  "recursive, r",
							Usage: "recurse into other conodes",
						},
					},
					Action: lsFetch,
				},
			},
		},
	}
	cliApp.Flags = []cli.Flag{
		app.FlagDebug,
		cli.StringFlag{
			Name:  "config, c",
			Value: "~/.config/scmgr/config.bin",
			Usage: "path to config-file",
		},
	}
	cliApp.Before = func(c *cli.Context) error {
		log.SetDebugVisible(c.Int("debug"))
		return nil
	}
	cliApp.Run(os.Args)
}

// Creates a new skipchain with the given roster
func create(c *cli.Context) error {
	log.Info("Create skipchain")
	group := readGroup(c, 0)
	client := skipchain.NewClient()
	data := []byte{}
	if address := c.String("url"); address != "" {
		if !strings.HasPrefix(address, "http") && !strings.HasPrefix(address, "config") {
			log.Fatal("Please give http- or config-address")
		}
		data = []byte(address)
	}
	sb, cerr := client.CreateGenesis(group.Roster, c.Int("base"), c.Int("height"),
		skipchain.VerificationStandard, data, nil)
	if cerr != nil {
		log.Fatal("while creating the genesis-roster:", cerr)
	}
	log.Infof("Created new skipblock with id %x", sb.Hash)
	cfg := getConfigOrFail(c)
	cfg.Sbb.AddBunch(sb)
	log.ErrFatal(cfg.save(c))
	return nil
}

// Joins a given skipchain
func join(c *cli.Context) error {
	log.Info("Joining skipchain")
	if c.NArg() < 2 {
		return errors.New("Please give group-file and id of known block")
	}
	group := readGroup(c, 0)
	client := skipchain.NewClient()
	hash, err := hex.DecodeString(c.Args().Get(1))
	if err != nil {
		return err
	}
	sbs, cerr := client.GetUpdateChain(group.Roster, hash)
	if cerr != nil {
		return cerr
	}
	latest := sbs[len(sbs)-1]
	genesis := latest.GenesisID
	if genesis == nil {
		genesis = latest.Hash
	}
	log.Infof("Joined skipchain %x", genesis)
	cfg := getConfigOrFail(c)
	cfg.Sbb.AddBunch(latest)
	log.ErrFatal(cfg.save(c))
	return nil
}

// Returns the number of calls.
func add(c *cli.Context) error {
	log.Info("Adding a block with a new group")
	if c.NArg() < 2 {
		return errors.New("Please give group-file and id to add")
	}
	group := readGroup(c, 1)
	cfg := getConfigOrFail(c)
	sb := cfg.Sbb.GetFuzzy(c.Args().First())
	if sb == nil {
		return errors.New("didn't find latest block - update first")
	}
	client := skipchain.NewClient()
	sbs, cerr := client.GetUpdateChain(sb.Roster, sb.Hash)
	if cerr != nil {
		return cerr
	}
	latest := sbs[len(sbs)-1]
	_, sbNew, cerr := client.AddSkipBlock(latest, group.Roster, nil)
	if cerr != nil {
		return errors.New("while storing block: " + cerr.Error())
	}
	cfg.Sbb.Store(sbNew)
	log.ErrFatal(cfg.save(c))
	log.Infof("Added new block %x to chain %x", sbNew.Hash, sbNew.GenesisID)
	return nil
}

// Updates a block to the latest block
func update(c *cli.Context) error {
	log.Info("Updating block")
	if c.NArg() < 1 {
		return errors.New("please give block-id to update")
	}
	cfg := getConfigOrFail(c)

	sb := cfg.Sbb.GetFuzzy(c.Args().First())
	if sb == nil {
		return errors.New("didn't find latest block in local store")
	}
	client := skipchain.NewClient()
	sbs, cerr := client.GetUpdateChain(sb.Roster, sb.Hash)
	if cerr != nil {
		return errors.New("while updating chain: " + cerr.Error())
	}
	if len(sbs) == 1 {
		log.Info("No new block available")
	} else {
		for _, b := range sbs[1:] {
			log.Infof("Adding new block %x to chain %x", b.Hash, b.GenesisID)
			cfg.Sbb.Store(b)
		}
	}
	latest := sbs[len(sbs)-1]
	log.Infof("Latest block of %x is %x", latest.GenesisID, latest.Hash)
	if c.Bool("data") {
		log.Info(base64.StdEncoding.EncodeToString(latest.Data))
	}
	log.ErrFatal(cfg.save(c))
	return nil
}

// lsKnown shows all known skipblocks
func lsKnown(c *cli.Context) error {
	cfg, err := loadConfig(c)
	if err != nil {
		return errors.New("couldn't read config: " + err.Error())
	}
	if len(cfg.Sbb.Bunches) == 0 {
		log.Info("Didn't find any blocks yet")
		return nil
	}
	genesis := sbl{}
	for _, bunch := range cfg.Sbb.Bunches {
		genesis = append(genesis, bunch.Latest)
	}
	sort.Sort(genesis)
	for _, g := range genesis {
		short := !c.Bool("long")
		if short {
			log.Infof("SkipChain %x", g.SkipChainID()[0:8])
		} else {
			log.Infof("SkipChain %x", g.SkipChainID())
		}
		sub := sbli{}
		for _, sb := range cfg.Sbb.GetBunch(g.SkipChainID()).SkipBlocks {
			sub = append(sub, sb)
		}
		sort.Sort(sub)
		for _, sb := range sub {
			log.Info("  " + sb.Sprint(short))
		}
	}
	return nil
}

func lsFetch(c *cli.Context) error {
	cfg := getConfigOrFail(c)
	rec := c.Bool("recursive")
	sisAll := map[network.ServerIdentityID]*network.ServerIdentity{}
	group := readGroup(c, 0)
	for _, bunch := range cfg.Sbb.Bunches {
		for _, sb := range bunch.SkipBlocks {
			for _, si := range sb.Roster.List {
				sisAll[si.ID] = si
			}
		}
	}
	for _, si := range group.Roster.List {
		sisAll[si.ID] = si
	}
	log.Info("The following ips will be searched:")
	for _, si := range sisAll {
		log.Info(si.Address)
	}
	client := skipchain.NewClient()
	sisNew := sisAll
	for len(sisNew) > 0 {
		sisIterate := sisNew
		sisNew = map[network.ServerIdentityID]*network.ServerIdentity{}
		for _, si := range sisIterate {
			log.Info("Fetching all skipchains from", si.Address)
			gasr, cerr := client.GetAllSkipchains(si)
			if cerr != nil {
				// Error is not fatal here - perhaps the node is down,
				// but we can continue anyway.
				log.Error(cerr)
				continue
			}
			for _, sb := range gasr.SkipChains {
				log.Printf("Found skipchain %x", sb.SkipChainID())
				cfg.Sbb.Store(sb)
				if rec {
					log.Print("rec")
					for _, si := range sb.Roster.List {
						if _, exists := sisAll[si.ID]; !exists {
							log.Print("Adding", si)
							sisNew[si.ID] = si
							sisAll[si.ID] = si
						}
					}
				}
			}
		}
	}
	return cfg.save(c)
}

// JSON skipblock element to be written in the index.html file
type jsonBlock struct {
	GenesisID string
	Servers   []string
	Data      []byte
}

// JSON list of skipblocks element to be written in the index.html file
type jsonBlockList struct {
	Blocks []jsonBlock
}

// sbl is used to make a nice output with ordered list of geneis-skipblocks.
type sbl []*skipchain.SkipBlock

func (s sbl) Len() int {
	return len(s)
}
func (s sbl) Less(i, j int) bool {
	return bytes.Compare(s[i].SkipChainID(), s[j].SkipChainID()) < 0
}
func (s sbl) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// sbli is used to make a nice output with ordered list of skipblocks of a
// skipchain.
type sbli sbl

func (s sbli) Len() int {
	return len(s)
}
func (s sbli) Less(i, j int) bool {
	return s[i].Index < s[j].Index
}
func (s sbli) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func readGroup(c *cli.Context, pos int) *app.Group {
	if c.NArg() <= pos {
		log.Fatal("Please give the group-file as argument")
	}
	name := c.Args().Get(pos)
	f, err := os.Open(name)
	log.ErrFatal(err, "Couldn't open group definition file")
	group, err := app.ReadGroupDescToml(f)
	log.ErrFatal(err, "Error while reading group definition file", err)
	if len(group.Roster.List) == 0 {
		log.ErrFatalf(err, "Empty entity or invalid group defintion in: %s",
			name)
	}
	return group
}

func getConfigOrFail(c *cli.Context) *config {
	cfg, err := loadConfig(c)
	log.ErrFatal(err)
	return cfg
}

func loadConfig(c *cli.Context) (*config, error) {
	cfgPath := app.TildeToHome(c.GlobalString("config"))
	_, err := os.Stat(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &config{
				Sbb: skipchain.NewSBBStorage(),
			}, nil
		}
		return nil, fmt.Errorf("Could not open file %s", cfgPath)
	}
	f, err := ioutil.ReadFile(cfgPath)
	if err != nil {
		return nil, err
	}
	_, cfg, err := network.Unmarshal(f)
	if err != nil {
		return nil, err
	}
	conf := cfg.(*config)
	if conf.Sbb == nil {
		conf.Sbb = skipchain.NewSBBStorage()
	}
	return conf, err
}

func (cfg *config) save(c *cli.Context) error {
	buf, err := network.Marshal(cfg)
	if err != nil {
		return err
	}
	file := app.TildeToHome(c.GlobalString("config"))
	cfgPath := path.Dir(file)
	_, err = os.Stat(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(cfgPath, 0770)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return ioutil.WriteFile(file, buf, 0660)
}