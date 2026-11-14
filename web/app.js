// TODO remove cursor grab feature for non-movable pieces. Don't even let it drag! (talking to you, blue!)

const svg = document.querySelector("#board");
const turnLabel = document.querySelector("#turn");
const statusLabel = document.querySelector("#status");
const rotateButton = document.querySelector("#rotate");
const resetButton = document.querySelector("#reset");
const sendButton = document.querySelector("#send");

const SVG_NS = "http://www.w3.org/2000/svg";
const CENTER = 380;
const SPACING = 64;
const SQRT3_OVER_2 = Math.sqrt(3) / 2;
const DIRECTIONS = [
  [0.5, 0.8660254],
  [1, 0],
  [0.5, -0.8660254],
  [-0.5, -0.8660254],
  [-1, 0],
  [-0.5, 0.8660254],
];

let socket;
let gameState;
let staged;
let drag;
let awaitingMove = false;

function keyOf(position) {
  return `${position.X},${position.Y}`;
}

function pointFor(position) {
  return {
    X: CENTER + (position.X - 4 + (position.Y - 4) * 0.5) * SPACING,
    Y: CENTER - (position.Y - 4) * SPACING * SQRT3_OVER_2,
  };
}

function svgElement(name, attributes = {}) {
  const element = document.createElementNS(SVG_NS, name);
  for (const [name, value] of Object.entries(attributes)) {
    element.setAttribute(name, value);
  }
  return element;
}

function getKingNeighbors({ KCoord, Ontop }) {
  return Ontop
    ? [KCoord, { X: KCoord.X + 1, Y: KCoord.Y }, { X: KCoord.X, Y: KCoord.Y + 1 }]
    : [KCoord, { X: KCoord.X - 1, Y: KCoord.Y }, { X: KCoord.X, Y: KCoord.Y - 1 }];
}

function triangleCenter(position) {
  const vertices = getKingNeighbors(position);
  const points = vertices.map(pointFor);
  return {
    X: points.reduce((sum, point) => sum + point.X, 0) / 3,
    Y: points.reduce((sum, point) => sum + point.Y, 0) / 3,
  };
}

function triangleMarkerPoints(position, scale) {
  const center = triangleCenter(position);
  return getKingNeighbors(position).map((vertex) => {
    const point = pointFor(vertex);
    return `${center.X + (point.X - center.X) * scale},${center.Y + (point.Y - center.Y) * scale}`;
  }).join(" ");
}

function setStatus(message, isError = false) {
  statusLabel.textContent = message;
  statusLabel.classList.toggle("error", isError);
}

function setControlsEnabled(enabled) {
  const canEdit = enabled && !awaitingMove;
  rotateButton.disabled = !canEdit || !staged || staged.type !== "move";
  resetButton.disabled = !canEdit || !staged;
  sendButton.disabled = !canEdit || !staged;
}

function addBoardBackground() {
  const corners = [
    { X: 0, Y: 4 }, { X: 0, Y: 8 }, { X: 4, Y: 8 },
    { X: 8, Y: 4 }, { X: 8, Y: 0 }, { X: 4, Y: 0 },
  ];
  svg.append(svgElement("polygon", {
    points: corners.map((corner) => {
      const point = pointFor(corner);
      return `${point.X},${point.Y}`;
    }).join(" "),
    fill: "#f0ecdf",
  }));
}

function drawGrid() {
  addBoardBackground();
  const positions = new Map(gameState.board.map((square) => [keyOf(square.position), square.position]));
  const directions = [[1, -1], [1, 0], [0, 1]];
  for (const position of positions.values()) {
    for (const [dx, dy] of directions) {
      const neighbor = { X: position.X + dx, Y: position.Y + dy };
      if (!positions.has(keyOf(neighbor))) continue;
      const start = pointFor(position);
      const end = pointFor(neighbor);
      svg.append(svgElement("line", {
        x1: start.X, y1: start.Y, x2: end.X, y2: end.Y, class: "grid-line",
      }));
    }
  }
  for (const position of positions.values()) {
    const point = pointFor(position);
    svg.append(svgElement("circle", {
      cx: point.X, cy: point.Y, r: 4.2, class: "vertex", "data-key": keyOf(position),
    }));
  }
}

function drawPiece(position, piece, isStaged = false) {
  const point = pointFor(position);
  const group = svgElement("g", {
    class: `piece ${piece.Col ? "true" : "false"}`,
    transform: `translate(${point.X} ${point.Y})`,
  });
  if (isStaged) {
    group.append(svgElement("circle", { r: 19, class: "staged-outline" }));
  }
  group.append(svgElement("circle", { r: 12 }));
  for (let angle = 0; angle < 6; angle++) {
    // Iterate over only the available angle of the piece.
    if (((piece.Ang>>angle)&1) == 0) {continue;}
    const [dx, dy] = DIRECTIONS[angle];
    group.append(svgElement("line", {
      x1: 0, y1: 0, x2: dx * 23, y2: dy * 23, class: "ray",
    }));
  }
  group.addEventListener("pointerdown", (event) => beginPieceDrag(event, group, position));
  svg.append(group);
}

function drawKing(position, color, isStaged = false) {
  const center = triangleCenter(position);
  const group = svgElement("g", {
    class: `king ${color ? "true" : "false"}`,
    transform: `translate(${center.X} ${center.Y})`,
  });
  group.append(svgElement("polygon", {
    points: triangleMarkerPoints(position, 0.46),
  }));
  if (isStaged) {
    group.append(svgElement("polygon", {
      points: triangleMarkerPoints(position, 0.68),
      class: "staged-outline",
    }));
  }
  const label = svgElement("text", { X: 0, Y: 1 });
  label.textContent = "K";
  group.append(label);
  group.addEventListener("pointerdown", (event) => beginKingDrag(event, group, position, color));
  svg.append(group);
}

function markTargets(targets) {
  for (const target of targets) {
    const vertex = svg.querySelector(`[data-key="${keyOf(target)}"]`);
    if (vertex) vertex.classList.add("legal-target");
  }
}

function clearTargets() {
  svg.querySelectorAll(".legal-target").forEach((vertex) => vertex.classList.remove("legal-target"));
}

function svgPoint(event) {
  const point = svg.createSVGPoint();
  point.x = event.clientX;
  point.y = event.clientY;
  return point.matrixTransform(svg.getScreenCTM().inverse());
}

function nearestDestination(targets, point, centerFor) {
  let nearest;
  let nearestDistance = Infinity;
  for (const target of targets) {
    const center = centerFor(target);
    const distance = Math.hypot(center.X - point.x, center.Y - point.y);
    if (distance < nearestDistance) {
      nearest = target;
      nearestDistance = distance;
    }
  }
  return nearestDistance <= SPACING * 0.52 ? nearest : undefined;
}

function beginDrag(event, node, options) {
  event.preventDefault();
  const start = svgPoint(event);
  drag = { node, start, baseTransform: node.getAttribute("transform"), options };
  markTargets(options.targets);
  node.setPointerCapture(event.pointerId);
}

function beginPieceDrag(event, node, position) {
  if (awaitingMove) return;
  const square = gameState.board.find((item) => keyOf(item.position) === keyOf(position));
  if (!square || square.piece.Ang === 0 || square.piece.Col !== gameState.turn) return;
  const targets = gameState.legalMoves[`${position.X},${position.Y}`];
  if (targets === 0) return;
  beginDrag(event, node, { type: "move", from: position, targets });
  node.addEventListener("pointermove", moveDrag);
  node.addEventListener("pointerup", finishDrag, { once: true });
  node.addEventListener("pointercancel", cancelDrag, { once: true });
}

function beginKingDrag(event, node, position, color) {
  if (awaitingMove) return;
  if (color !== gameState.turn) return;
  const currentKing = color ? gameState.trueKing : gameState.falseKing;
  if (currentKing.map(keyOf).sort().join("|") !== position.map(keyOf).sort().join("|")) return;
  const targets = gameState.legalKingMoves;
  if (targets.length === 0) return;
  for (const target of targets) {
    const center = triangleCenter(target);
    svg.append(svgElement("circle", {
      cx: center.X, cy: center.Y, r: 14, class: "king-target",
    }));
  }
  beginDrag(event, node, { type: "king_move", targets });
  node.addEventListener("pointermove", moveDrag);
  node.addEventListener("pointerup", finishDrag, { once: true });
  node.addEventListener("pointercancel", cancelDrag, { once: true });
}

function moveDrag(event) {
  if (!drag) return;
  const current = svgPoint(event);
  drag.node.setAttribute(
    "transform",
    `${drag.baseTransform} translate(${current.x - drag.start.x} ${current.y - drag.start.y})`,
  );
}

function finishDrag(event) {
  if (!drag) return;
  const current = svgPoint(event);
  const activeDrag = drag;
  drag = undefined;
  activeDrag.node.removeEventListener("pointermove", moveDrag);
  activeDrag.node.removeEventListener("pointercancel", cancelDrag);
  clearTargets();
  svg.querySelectorAll(".king-target").forEach((target) => target.remove());
  const destination = nearestDestination(
    activeDrag.options.targets,
    current,
    activeDrag.options.type === "move" ? pointFor : triangleCenter,
  );
  if (!destination) {
    activeDrag.node.setAttribute("transform", activeDrag.baseTransform);
    setStatus("Drop on a highlighted legal destination.");
    return;
  }
  staged = activeDrag.options.type === "move"
    ? { type: "move", from: activeDrag.options.from, to: destination, rotation: 0 }
    : { type: "king_move", to: destination };
  render();
  setStatus("Move staged. Rotate, send, or reset.");
}

function cancelDrag() {
  if (!drag) return;
  const node = drag.node;
  const baseTransform = drag.baseTransform;
  drag = undefined;
  node.removeEventListener("pointermove", moveDrag);
  node.removeEventListener("pointerup", finishDrag);
  clearTargets();
  svg.querySelectorAll(".king-target").forEach((target) => target.remove());
  node.setAttribute("transform", baseTransform);
}

/**
 * Rotates a number left by x bits, treating it as a 6-bit value.
 * Bits 6 and 7 are forced to zero.
 */
function rotateLeft6Bit(b, x) {
    // Mask to keep only the lower 6 bits (bits 6 and 7 set to 0)
    b &= 0x3F;

    // Normalize rotation amount to be within 0-5 (handles negative shifts safely)
    x = ((x % 6) + 6) % 6;

    if (x === 0) {
        return b;
    }

    // Perform the 6-bit left rotation using unsigned right shift (>>>) and re-apply the mask
    return ((b << x) | (b >>> (6 - x))) & 0x3F;
}

function stagedPiece() {
  if (!staged || staged.type !== "move") return undefined;
  const square = gameState.board.find((item) => keyOf(item.position) === keyOf(staged.from));
  if (!square) return undefined;
  const angles = rotateLeft6Bit(square.piece.Ang, staged.rotation);
  return { Col: square.piece.Col, Ang: angles };
}

function render() {
  if (!gameState) return;
  svg.replaceChildren();
  drawGrid();
  for (const square of gameState.board) {
    if (square.piece.Ang === 0) continue;
    if (staged?.type === "move" &&
        (keyOf(square.position) === keyOf(staged.from) || keyOf(square.position) === keyOf(staged.to))) {
      continue;
    }
    drawPiece(square.position, square.piece);
  }
  if (staged?.type === "move") {
    const piece = stagedPiece();
    if (piece) drawPiece(staged.to, piece, true);
  }
  const trueKing = staged?.type === "king_move" && gameState.turn
    ? staged.to
    : gameState.trueKing;
  const falseKing = staged?.type === "king_move" && !gameState.turn
    ? staged.to
    : gameState.falseKing;
  drawKing(trueKing, true, staged?.type === "king_move" && gameState.turn);
  drawKing(falseKing, false, staged?.type === "king_move" && !gameState.turn);
  turnLabel.textContent = gameState.turn ? "Blue to move" : "Orange to move";
  setControlsEnabled(socket?.readyState === WebSocket.OPEN);
}

function onMessage(event) {
  let message;
  try {
    message = JSON.parse(event.data);
  } catch {
    setStatus("The game server sent an unreadable message.", true);
    return;
  }
  console.log(message);
  if (message.type === "error") {
    awaitingMove = false;
    setControlsEnabled(socket?.readyState === WebSocket.OPEN);
    setStatus(message.message, true);
    return;
  }
  if (message.type !== "state") return;
  if (gameState && message.revision < gameState.revision) return;
  const revisionChanged = !gameState || message.revision !== gameState.revision;
  awaitingMove = false;
  gameState = message;
  if (revisionChanged) staged = undefined;
  render();
  if (revisionChanged) {
    setStatus(`${gameState.turn ? "Blue" : "Orange"} to move.`);
  } else if (!staged) {
    setStatus("Drag a piece or king to a highlighted destination.");
  }
}

function connect() {
  const scheme = location.protocol === "https:" ? "wss:" : "ws:";
  socket = new WebSocket(`${scheme}//${location.host}/ws`);
  socket.addEventListener("open", () => {
    setStatus("Connected. Waiting for the current game position…");
  });
  socket.addEventListener("message", onMessage);
  socket.addEventListener("close", () => {
    setStatus("Disconnected from the local game server. Refresh to reconnect.", true);
    setControlsEnabled(false);
  });
  socket.addEventListener("error", () => {
    setStatus("Could not connect to the local game server.", true);
  });
}

rotateButton.addEventListener("click", () => {
  if (!staged || staged.type !== "move") return;
  staged.rotation = (staged.rotation + 1) % 6;
  render();
  setStatus("Piece rotated. Send the move or reset it.");
});

resetButton.addEventListener("click", () => {
  staged = undefined;
  render();
  setStatus("Staged move reset.");
});

sendButton.addEventListener("click", () => {
  if (!staged || socket?.readyState !== WebSocket.OPEN) return;
  try {
    if (staged.type === "move") {
      socket.send(JSON.stringify({
        type: "move",
        from: staged.from,
        to: staged.to,
        rotation: staged.rotation,
      }));
    } else {
      socket.send(JSON.stringify({ type: "king_move", king: staged.to }));
    }
  } catch (error) {
    setStatus(`Could not send move: ${error.message}`, true);
    setControlsEnabled(socket?.readyState === WebSocket.OPEN);
    return;
  }
  awaitingMove = true;
  setControlsEnabled(false);
  setStatus("Move sent. Waiting for the server…");
});

connect();
