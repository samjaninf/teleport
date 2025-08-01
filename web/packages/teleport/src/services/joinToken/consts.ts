/*
 * Teleport
 * Copyright (C) 2025 Gravitational, Inc.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

import { ListJoinTokensResponse } from './types';

export function validateListJoinTokensResponse(
  data: unknown
): data is ListJoinTokensResponse {
  if (typeof data !== 'object' || data === null) {
    return false;
  }

  if (!('items' in data)) {
    return false;
  }

  if (!Array.isArray(data.items)) {
    return false;
  }

  return data.items.every(x => typeof x === 'object' || x !== null);
}
