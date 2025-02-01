// Teleport
// Copyright (C) 2023  Gravitational, Inc.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

fn main() {
    // the cwd of build scripts is the root of the crate.
    let bindings = cbindgen::Builder::new()
        .with_crate(".")
        .with_language(cbindgen::Language::C)
        .generate()
        .unwrap();

    // atomically swap the header in place, just in case there's multiple
    // compilations at the same time.
    let out = tempfile::NamedTempFile::new_in(".").unwrap();
    bindings.write(&out);

    // TODO(espadolini): target-specific paths. Ideally we want the header in
    // target/arch-vendor-os/release/ so that the Go side can refer to it like
    // it refers to the .a, but the only way I found so far is to use
    // ${OUT_DIR}/../../ which isn't really guaranteed to work, afaict.
    out.persist("librdprs.h").unwrap();
}
