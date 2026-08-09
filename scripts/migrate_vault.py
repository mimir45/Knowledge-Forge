#!/usr/bin/env python3
"""One-time vault migration: pre-Forge topology -> DESIGN §7 topology.

Throwaway tooling. This never ships in the `forge` binary (CLAUDE.md, "Only surviving
Python"). It does exactly the two things Go cannot do safely on its own -- move files and
infer `type` from where a note currently lives -- and leaves the whole contract shape
(dates, tag aliases, key order, sources) to `forge validate --all --fix`, which is tested.

Safety contract, in the order it is enforced:
  1. --dry-run is the DEFAULT. Writing requires an explicit --apply.
  2. --apply refuses a dirty vault git tree; the migration is irreversible and `git
     checkout .` is the only undo.
  3. --backup copies the whole vault before anything is touched.
  4. The complete old->new map is built and checked for collisions BEFORE any write.
     Two notes landing on one path would be a silent overwrite on a vault with no
     backups, so a collision aborts the run.
  5. Body content is never reordered, rewritten or deleted. The only edit inside a
     body is retargeting a `[[wikilink]]` whose destination this migration moves, and a
     bare `[[name]]` is retargeted only when its basename is unique in the vault.
  6. Every inferred field is stamped `confidence: low`.
"""

import argparse
import os
import re
import shutil
import subprocess
import sys

# Directories that are ingest input or history, not vault notes (taxonomy.md §5).
# They stay exactly where they are: `source:` values point into sources/daily/.
NON_NOTE_DIRS = ("raw/", "sources/", "archive/", "_archive/", ".git/", ".obsidian/")
NON_NOTE_FILES = {"CLAUDE.md", "README.md", "lint-report.md", "index.md", "log.md"}

# taxonomy.md §4, "Mapping the old topology onto `type`".
DIR_TYPE = {
    "decisions": ("decision", "high"),
    "concepts": ("concept", "low"),
    "entities": ("concept", "low"),
    "syntheses": ("concept", "low"),
}
OUTAGE = re.compile(r"\b(outage|incident|down|postmortem|post-mortem)\b", re.I)
IMPERATIVE = re.compile(r"^(?:\d+\.\s|```|\$ |- \[ \])", re.M)


# TRANSLIT mirrors pkg/vault.translit exactly. Without it `[^a-z0-9]` deletes every
# accented letter, so `BÖLÜM I: TEORİ TEMELLERİ` slugifies to `b-l-m-i-teori-temelleri`
# and the migration writes that unreadable name to disk. The rename is irreversible and
# `forge slug` would disagree with it forever, so the two tables have to stay in step.
TRANSLIT = {
    "ç": "c", "ğ": "g", "ı": "i", "ö": "o", "ş": "s", "ü": "u",
    "á": "a", "à": "a", "â": "a", "ä": "a", "å": "a", "ã": "a",
    "é": "e", "è": "e", "ê": "e", "ë": "e",
    "í": "i", "ì": "i", "î": "i", "ï": "i",
    "ó": "o", "ò": "o", "ô": "o", "õ": "o",
    "ú": "u", "ù": "u", "û": "u",
    "ñ": "n", "ß": "ss", "æ": "ae", "ø": "o",
    "+": "plus", "&": "and", "#": "sharp",
}
MAX_SLUG_LEN = 80


def slugify(text):
    folded = "".join(TRANSLIT.get(c, c) for c in text.lower())
    s = re.sub(r"-+", "-", re.sub(r"[^a-z0-9]+", "-", folded)).strip("-")
    return s if len(s) <= MAX_SLUG_LEN else truncate_at_boundary(s)


def truncate_at_boundary(s):
    """Mirrors pkg/vault.truncateAtBoundary: cut to MAX_SLUG_LEN without leaving a
    trailing partial word. The slug is also the destination filename, so a slug the
    schema rejects would be baked into a path this migration cannot take back."""
    s = s[:MAX_SLUG_LEN]
    cut = s.rfind("-")
    return (s[:cut] if cut > 0 else s).strip("-")


def split_frontmatter(text):
    """Return (yaml_source, body). Mirrors pkg/vault.SplitFrontmatter."""
    m = re.match(r"\ufeff?---\r?\n(.*?)\r?\n---\r?\n", text, re.S)
    if not m:
        return "", text
    return m.group(1), text[m.end():]


def fm_value(fm_src, key):
    m = re.search(r"^%s:\s*(.*)$" % re.escape(key), fm_src, re.M)
    return m.group(1).strip().strip("\"'") if m else ""


def fm_list(fm_src, key):
    """Read a list value in either the inline `key: [a, b]` or the block `- a` form.
    The vault uses both, and taxonomy.md §1 blames that split for its own tag counts
    disagreeing with the audit's."""
    inline = fm_value(fm_src, key)
    if inline.startswith("["):
        items = inline.strip("[]").split(",")
    else:
        m = re.search(r"^%s:[ \t]*\n((?:[ \t]*-[ \t]+.*\n?)+)" % re.escape(key), fm_src, re.M)
        items = re.findall(r"^[ \t]*-[ \t]+(.*)$", m.group(1), re.M) if m else []
    return [i.strip().strip("\"'").lower() for i in items if i.strip()]


def load_stack_vocab():
    """The closed `stack` list and its alias map, read from the one source of truth.
    Hard-coding 41 values here would let the script and the schema drift apart."""
    text = read(os.path.join(os.path.dirname(os.path.abspath(__file__)),
                             os.pardir, "references", "schema.yaml"))
    values = re.search(r"\n  stack:\n(.*?)\n  tags:\n", text, re.S)
    amap = re.search(r"\naliases:\n  stack:\n(.*?)\n  tags:\n", text, re.S)
    if not values or not amap:
        die("references/schema.yaml: cannot read the stack vocabulary")
    return (set(re.findall(r"^      - ([a-z0-9-]+)$", values.group(1), re.M)),
            dict(re.findall(r"^    ([a-z0-9-]+): ([a-z0-9-]+)$", amap.group(1), re.M)))


def title_of(fm_src, body, rel):
    for candidate in (fm_value(fm_src, "title"), first_h1(body)):
        if candidate:
            return candidate
    return os.path.splitext(os.path.basename(rel))[0].replace("-", " ")


def first_h1(body):
    m = re.search(r"^#\s+(.+)$", body, re.M)
    return m.group(1).strip() if m else ""


def is_content_note(rel):
    if not rel.endswith(".md") or rel.startswith(NON_NOTE_DIRS):
        return False
    base = os.path.basename(rel)
    return base not in NON_NOTE_FILES and not base.startswith("_index")


def walk(root):
    out = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames[:] = [d for d in dirnames if d not in (".git", ".obsidian", ".forge")]
        for name in sorted(filenames):
            if name.endswith(".md"):
                out.append(os.path.relpath(os.path.join(dirpath, name), root))
    return sorted(out)


def infer_type(rel, body):
    """Infer `type` from location, then from body cues. Returns (type, confidence)."""
    top = rel.split("/")[0]
    if top in DIR_TYPE:
        return DIR_TYPE[top]
    if top == "issues":
        return ("incident" if OUTAGE.search(body) else "pitfall"), "low"
    if top == "TIL":
        return ("howto" if IMPERATIVE.search(body) else "concept"), "low"
    return "concept", "low"


def infer_depth(rel):
    """syntheses/ notes are cross-cutting by construction (taxonomy.md §4)."""
    return "4" if rel.split("/")[0] == "syntheses" else ""


class Plan:
    """The complete set of changes, computed before a single byte is written."""

    def __init__(self):
        self.moves = {}       # old rel -> new rel
        self.fields = {}      # old rel -> {key: value} to inject
        self.skipped = []     # (rel, reason) for every .md left in place
        self.collisions = []  # (new rel, [old rels]) -- any entry aborts the run


def build_plan(root):
    plan, claimed = Plan(), {}
    vocab = load_stack_vocab()
    for rel in walk(root):
        if not is_content_note(rel):
            plan.skipped.append((rel, skip_reason(rel)))
            continue
        add_note(root, rel, plan, claimed, vocab)
    # Two notes claiming one path is one way to lose a note; a path that already exists on
    # disk is the other, and only this second check catches a half-applied earlier run.
    plan.collisions = [(dst, srcs) for dst, srcs in sorted(claimed.items())
                       if len(srcs) > 1 or os.path.exists(os.path.join(root, dst))]
    return plan


def skip_reason(rel):
    if rel.startswith(("raw/", "sources/")):
        return "ingest input, not a note (taxonomy.md §5)"
    if rel.startswith(("archive/", "_archive/")):
        return "history, left in place"
    return "non-note file"


def add_note(root, rel, plan, claimed, vocab):
    text = read(os.path.join(root, rel))
    fm_src, body = split_frontmatter(text)
    ntype, conf = infer_type(rel, body)
    slug = fm_value(fm_src, "slug") or slugify(title_of(fm_src, body, rel))
    dst = "notes/%s/%s.md" % (ntype, slug)
    plan.moves[rel] = dst
    plan.fields[rel] = note_fields(fm_src, rel, ntype, conf, slug, vocab)
    claimed.setdefault(dst, []).append(rel)


def note_fields(fm_src, rel, ntype, conf, slug, vocab):
    """Only fields the migration is entitled to write. Everything else is --fix's job."""
    out = {}
    if not fm_value(fm_src, "type"):
        out["type"] = ntype
        # confidence records how the type was chosen, so it tracks that inference and
        # nothing else. `decisions/` maps at high confidence; every other rule guesses.
        if not fm_value(fm_src, "confidence"):
            out["confidence"] = conf
    if not fm_value(fm_src, "slug"):
        out["slug"] = slug
    if depth := infer_depth(rel):
        if not fm_value(fm_src, "depth"):
            out["depth"] = depth
    # schema.yaml, origin: "`import` is what the Phase 1 migration stamps on every
    # pre-existing note." A documented constant, not an inference -- no confidence.
    if not fm_value(fm_src, "origin"):
        out["origin"] = "import"
    out.update(axis_fields(fm_src, vocab))
    return out


def axis_fields(fm_src, vocab):
    """Partition the v1 single `tags:` list onto the contract's two axes (taxonomy.md §1:
    technologies are *promoted* to `stack`, concepts stay tags). This redistributes values
    the user already wrote; it invents nothing. Canonicalize first or `spring` would stay
    a tag instead of becoming `spring-boot`. An empty side is left absent rather than
    filled, so `forge validate` keeps reporting it as a note a human has to name.
    """
    values, aliases = vocab
    legacy = fm_list(fm_src, "tags")
    if fm_list(fm_src, "stack"):
        return {}  # already on the contract's two axes; leave it alone
    stack, tags = [], []
    for item in legacy:
        canon = aliases.get(item, item)
        bucket = stack if canon in values else tags
        if canon not in bucket:
            bucket.append(canon)
    # Never truncated to max_items: dropping a value silently is the loss this whole
    # script exists to avoid. An over-long list is reported by `forge validate` instead.
    out = {k: "[%s]" % ", ".join(v) for k, v in (("stack", stack), ("tags", tags)) if v}
    if legacy and "tags" not in out:
        # Every old tag was a technology and moved to `stack`. The key has to be written
        # back empty, because inject only replaces keys the split names: leaving it out
        # would leave the old list in place and put the same values on both axes.
        out["tags"] = "[]"
    return out


# ---------------------------------------------------------------- link rewriting

# A path-qualified link names a directory that is about to change, so it always has to be
# retargeted. A bare [[name]] resolves by basename (AUDIT §11) and normally survives a
# move untouched -- but not when the destination slug differs from the old filename, and
# the real vault has a six-note series that cross-links entirely in the bare form.
WIKILINK = re.compile(r"\[\[([^\]|#]+/[^\]|#]+?)(\.md)?([#|][^\]]*)?\]\]")


def link_map(moves):
    """old link target (path, no extension) -> new link target."""
    return {strip_md(old): strip_md(new) for old, new in moves.items()}


def bare_map(root, moves):
    """old basename -> new basename, for the renames only.

    A basename that is not unique in the vault is excluded: resolving it would mean
    guessing which of two notes a bare link meant, and guessing wrong rewrites a link to
    point at the wrong note -- worse than leaving it dangling, because it looks fine.
    """
    seen = {}
    for rel in walk(root):
        seen[base(rel)] = seen.get(base(rel), 0) + 1
    return {base(o): base(n) for o, n in moves.items()
            if base(o) != base(n) and seen.get(base(o), 0) == 1}


def strip_md(rel):
    return rel[:-3] if rel.endswith(".md") else rel


def rewrite_links(text, lmap, bmap):
    """Returns (new_text, count). Never touches anything outside the [[...]] span.

    Path-qualified first: its output contains a slash, so BARE cannot then match and
    rewrite the same link twice.
    """
    text, a = sub_targets(WIKILINK, text, lmap)
    text, b = sub_targets(BARE, text, bmap)
    return text, a + b


def sub_targets(pattern, text, tmap):
    count = 0

    def sub(m):
        nonlocal count
        target = tmap.get(m.group(1).strip())
        if target is None:
            return m.group(0)
        count += 1
        return "[[%s%s]]" % (target, m.group(3) or "")

    return pattern.sub(sub, text), count


# ---------------------------------------------------------------- frontmatter injection

def drop_keys(fm_src, keys):
    """Remove top-level keys and any block-form items under them. Called only for keys
    `fields` is about to write, so nothing is deleted without being replaced."""
    out, skipping = [], False
    for line in fm_src.split("\n"):
        if line[:1].strip():  # a top-level key line, not an indented continuation
            skipping = line.split(":", 1)[0] in keys
        if not skipping:
            out.append(line)
    return "\n".join(out)


def inject(text, fields):
    """Add keys to the frontmatter block, creating one if the note has none.

    A key the fields dict writes replaces the old one rather than joining it: `tags` is
    rewritten by the two-axis split, and two `tags:` keys in one mapping is a YAML error
    Go would refuse to load. Every other key is added only when absent, and the body is
    never reordered -- `forge validate --fix` owns key order and is the tested one.
    """
    if not fields:
        return text
    added = "".join("%s: %s\n" % (k, v) for k, v in fields.items())
    fm_src, body = split_frontmatter(text)
    if fm_src:
        return "---\n%s\n%s---\n%s" % (drop_keys(fm_src, fields), added, body)
    if text.lstrip("﻿").startswith("---"):
        return text  # an unterminated block: prepending a second one would compound it
    # No blank line after the closing fence: it would prepend a byte to the body of a
    # note that had no frontmatter, and "body is copied through byte-for-byte" is the
    # property this whole migration rests on. pkg/vault.marshalDoc emits the same shape.
    return "---\n%s---\n%s" % (added, text)


# ---------------------------------------------------------------- git / io

def read(path):
    with open(path, encoding="utf-8") as fh:
        return fh.read()


def git(root, *args):
    return subprocess.run(["git", "-C", root, *args],
                          capture_output=True, text=True, check=False)


def require_clean_tree(root):
    if not os.path.isdir(os.path.join(root, ".git")):
        die("%s is not a git repository; refusing to migrate without an undo path" % root)
    res = git(root, "status", "--porcelain")
    if res.returncode != 0:
        die("git status failed: %s" % res.stderr.strip())
    if res.stdout.strip():
        die("vault git tree is dirty; commit or stash first:\n%s" % res.stdout.rstrip())


def die(msg):
    print("migrate_vault: %s" % msg, file=sys.stderr)
    sys.exit(2)


def make_backup(root, dest):
    if os.path.exists(dest):
        die("backup destination %s already exists" % dest)
    shutil.copytree(root, dest, symlinks=True)
    print("backup: %s -> %s" % (root, dest))


# ---------------------------------------------------------------- apply

def apply_plan(root, plan):
    maps = link_map(plan.moves), bare_map(root, plan.moves)
    links = rewrite_all(root, plan, maps)
    for old, new in sorted(plan.moves.items()):
        move(root, old, new)
    for name in ("moc", "_inbox", "_archive", "profiles"):
        os.makedirs(os.path.join(root, name), exist_ok=True)
    return links


def rewrite_all(root, plan, maps):
    """Rewrites links and injects fields in one pass, before anything moves."""
    total = 0
    for rel in walk(root):
        path = os.path.join(root, rel)
        text = read(path)
        new_text, n = rewrite_links(text, *maps)
        new_text = inject(new_text, plan.fields.get(rel, {}))
        total += n
        if new_text != text:
            with open(path, "w", encoding="utf-8") as fh:
                fh.write(new_text)
    return total


def move(root, old, new):
    src, dst = os.path.join(root, old), os.path.join(root, new)
    if os.path.exists(dst):
        die("refusing to overwrite %s (this should have been caught as a collision)" % new)
    os.makedirs(os.path.dirname(dst), exist_ok=True)
    if git(root, "mv", old, new).returncode != 0:
        shutil.move(src, dst)


# ---------------------------------------------------------------- reporting

def report(root, plan, verbose):
    maps = link_map(plan.moves), bare_map(root, plan.moves)
    links = count_links(root, maps)
    inferred = sum(1 for f in plan.fields.values() if f.get("confidence") == "low")
    print("\n%s" % ("=" * 68))
    print("vault:              %s" % root)
    print("files to move:      %d" % len(plan.moves))
    print("links to rewrite:   %d  (path-qualified, plus bare links to renamed notes)" % links)
    print("fields inferred:    %d notes stamped confidence: low" % inferred)
    print("left in place:      %d" % len(plan.skipped))
    print("collisions:         %d" % len(plan.collisions))
    print("=" * 68)
    print_types(plan)
    print_axes(plan)
    print_bare_breaks(root, plan)
    print_skipped(plan)
    if verbose:
        print_moves(plan)
    print_collisions(plan)


# A bare [[name]] carries no directory, so it survives a move by basename resolution --
# but not a rename, and the destination slug is not always the old filename.
BARE = re.compile(r"\[\[([^\]|#/]+?)(\.md)?([#|][^\]]*)?\]\]")


def base(rel):
    return os.path.splitext(os.path.basename(rel))[0]


def print_bare_breaks(root, plan):
    """Bare links this migration renames but cannot safely retarget -- the basename is
    not unique, so which note the link meant is a guess. These break, and a human has to
    say which one was intended."""
    safe = bare_map(root, plan.moves)
    renamed = {base(o) for o, n in plan.moves.items() if base(o) != base(n)}
    hits = {}
    for rel in walk(root):
        for m in BARE.finditer(read(os.path.join(root, rel))):
            target = m.group(1).strip()
            if target in renamed and target not in safe:
                hits.setdefault(target, []).append(rel)
    print("\nbare [[name]] links that break (ambiguous basename, not retargeted): %d"
          % sum(len(v) for v in hits.values()))
    for target, srcs in sorted(hits.items()):
        print("  [[%s]]  <- %s" % (target, ", ".join(sorted(set(srcs)))))


def print_axes(plan):
    """The two-axis split's residue. These are the notes the migration cannot finish on
    its own: `stack` and `tags` are both required with min_items 1, so a note whose old
    tags fell entirely on one side stays invalid until a human names the other."""
    n = len(plan.fields)
    counts = {"no stack value in its old tags": 0, "nothing left for tags": 0,
              "stack over max_items 6": 0, "tags over max_items 8": 0}
    for f in plan.fields.values():
        counts["no stack value in its old tags"] += not items_of(f, "stack")
        counts["nothing left for tags"] += not items_of(f, "tags")
        counts["stack over max_items 6"] += len(items_of(f, "stack")) > 6
        counts["tags over max_items 8"] += len(items_of(f, "tags")) > 8
    print("\ntags -> stack + tags split (taxonomy.md §1), over %d notes:" % n)
    for reason, c in counts.items():
        print("  %-32s %3d  %s" % (reason, c, "needs a human" if c else "-"))


def items_of(fields, key):
    return [i for i in fields.get(key, "").strip("[]").split(",") if i.strip()]


def count_links(root, maps):
    return sum(rewrite_links(read(os.path.join(root, rel)), *maps)[1]
               for rel in walk(root))


def print_types(plan):
    counts = {}
    for dst in plan.moves.values():
        counts[dst.split("/")[1]] = counts.get(dst.split("/")[1], 0) + 1
    print("\ntype distribution (inferred from old location):")
    for t, n in sorted(counts.items(), key=lambda kv: (-kv[1], kv[0])):
        print("  notes/%-10s %3d" % (t + "/", n))


def print_skipped(plan):
    """Named explicitly: '109 -> 93 moved' with no account of the other 16 reads as
    data loss, and the approval decision should not require that inference."""
    groups = {}
    for rel, reason in plan.skipped:
        groups.setdefault(reason, []).append(rel)
    print("\nleft in place (not notes -- nothing is dropped):")
    for reason, rels in sorted(groups.items()):
        print("  %-45s %3d  e.g. %s" % (reason, len(rels), rels[0]))


def print_moves(plan):
    print("\nmoves:")
    for old, new in sorted(plan.moves.items()):
        print("  %-52s -> %s" % (old, new))


def print_collisions(plan):
    if not plan.collisions:
        return
    print("\nCOLLISIONS -- two notes claim one destination path. Refusing to run:")
    for dst, srcs in plan.collisions:
        print("  %s" % dst)
        for s in srcs:
            print("      <- %s" % s)


def main():
    ap = argparse.ArgumentParser(description=__doc__,
                                 formatter_class=argparse.RawDescriptionHelpFormatter)
    ap.add_argument("vault")
    ap.add_argument("--apply", action="store_true",
                    help="actually write. Without it this is a dry run.")
    ap.add_argument("--backup", metavar="DIR", help="copy the vault here before writing")
    ap.add_argument("--verbose", action="store_true", help="print every move")
    args = ap.parse_args()

    root = os.path.abspath(args.vault)
    plan = build_plan(root)
    report(root, plan, args.verbose)
    if plan.collisions:
        die("%d destination collision(s); nothing was written" % len(plan.collisions))
    if not args.apply:
        print("\nDRY RUN -- nothing written. Re-run with --apply to commit these changes.")
        return
    require_clean_tree(root)
    if args.backup:
        make_backup(root, os.path.abspath(args.backup))
    links = apply_plan(root, plan)
    print("\nAPPLIED: %d files moved, %d links rewritten." % (len(plan.moves), links))
    print("Next: forge validate --all --fix, then commit the vault.")


if __name__ == "__main__":
    main()
