# 📋 Project Spec: AI-Powered Sum-100 Puzzle (Agentic Game) v3.0

โปรเจกต์เกมแนว Logic Puzzle ที่ใช้ AI (WebLLM - Client-side LLM) เป็นตัวขับเคลื่อนเกมผ่านระบบ Function Calling โดยผู้เล่นสามารถโต้ตอบกับเกมผ่านการพิมพ์คำสั่งภาษาธรรมชาติ

---

## 🎮 1. Game Overview

| Item | Description |
|------|-------------|
| **Concept** | เติมตัวเลขลงในตาราง (Grid) โดยผลรวมของทุกแถว (Row) และทุกคอลัมน์ (Column) ต้องเท่ากับ **100** |
| **Grid Size** | เริ่มต้นที่ 5x5 (สามารถขยายเป็น NxN ได้ในอนาคต โดยแถว = คอลัมน์เสมอ) |
| **Pre-filled Cells** | 9 ช่อง จาก 25 ช่อง (เหลือ 16 ช่องให้ผู้เล่นเติม) |
| **Number Range** | 1-99 เท่านั้น (ไม่มีเลข 0, ไม่ติดลบ) |
| **Target** | ทุกแนวตั้งและแนวนอนรวมกันได้ 100 (ไม่นับแนวทแยง) |

---

## 🕹️ 2. Core Game Mechanics & Rules

### Game Flow (Updated)
1. **Initial Load** - User เข้าเกมครั้งแรก:
   - Backend ส่ง puzzles ที่ยังไม่ complete ให้ทั้งหมดสูงสุด 20 ใบ
   - Frontend เก็บและ render ให้ user ดูได้ทั้งหมด
   - User สามารถกด next/back เพื่อเลือกเล่น puzzle แต่ละใบ

2. **Playing** - User เล่น puzzle:
   - User ใส่เลขลงในช่องว่าง
   - สามารถใส่คำตอบได้ทีละหลายช่อง
   - กด submit เพื่อส่งคำตอบไปตรวจสอบ

3. **Verification** - ตรวจสอบคำตอบ:
   - Backend ตรวจทุกช่องที่ส่งมา
   - แสดง feedback ทันที (CORRECT, TOO_LOW, TOO_HIGH)
   - นับ mistakes ทุกช่องที่ผิด

4. **Completion** - เสร็จ puzzle:
   - เมื่อทุกช่องถูกต้อง → Puzzle Complete
   - Backend mark puzzle ว่า complete ให้ guest
   - User สามารถเล่น puzzle อื่นที่ยังไม่ complete ได้

### Ad Reward System (Future Feature)
- เมื่อ user complete 20 ด่านแล้ว → ไม่มี puzzle ให้เล่น
- Frontend แสดงปุ่มดูโฆษณา
- หลังดูโฆษณา → Frontend เรียก `getAvailablePuzzles` อีกครั้ง
- Backend จะส่ง puzzles ใหม่ อีก 20 ใบ (จากที่เหลือใน puzzle_pool)

### Validation Logic & Feedback
- หากเลขในช่องใดผิด ระบบจะตอบกลับเป็น **รายช่อง**:
  - ✅ **Correct** - ถูกต้อง
  - ⬆️ **Too High** - มากเกินไป
  - ⬇️ **Too Low** - น้อยเกินไป
- ระบบนับ **mistakes ทุกช่องที่ส่งมาผิด** (เช่น ส่ง 4 ช่อง ผิด 4 ช่อง = 4 mistakes)

### Security Note
- **Solution เก็บอยูใน backend เท่านั้น**
- Frontend ไม่มีทางรู้คำตอบ
- Validate ทุก submission บน backend
- Frontend ไม่สามารถ hack หา solution ได้

---

## 🏆 3. Leaderboard System

| Item | Description |
|------|-------------|
| **Scope** | Global Leaderboard (ผู้เล่นทุกคนอยู่ board เดียวกัน) |
| **Ranking Metric** | **Total Mistakes** เท่านั้น (ยิ่งน้อยยิ่งดี) |
| **Time Tracking** | ไม่มี (ไม่วัดความเร็ว) |
| **Display** | Top 100 players |

### Authentication
- **ไม่มีระบบ Login**
- เมื่อจบเกม ผู้เล่นสามารถ:
  - ใส่ **Username** เพื่อบันทึกลง Leaderboard
  - หรือเล่นเป็น **Guest** (ใช้ UUID จาก client)

---

## 🤖 4. AI & Agentic Features (WebLLM - Client-side)

> **⚠️ PHASE 2 FEATURE** - ฟีเจอร์นี้จะทำหลังจากเกมหลักทำเสร็จสมบูรณ์แล้ว

### Function Calling (Client-side)
| Function | Description |
|----------|-------------|
| `fill_cells([{row, col, value}])` | รับคำสั่งจาก User เพื่อเติมเลขในตาราง (รองรับหลายช่องพร้อมกัน) |
| `verify_grid()` | เรียกใช้งานเมื่อ User สั่งให้ตรวจคำตอบผ่าน Chat |
| `get_current_state()` | ดึงสถานะ grid ปัจจุบันเพื่อวิเคราะห์ |

### Consultative Chat
- ผู้เล่นสามารถพิมพ์ถามแนวทางได้ เช่น:
  - *"Row 1 should have what number?"*
  - *"แถวแรกควรใส่เลขอะไรดี?"*
  - *"1行目に何を入れればいい？"*
- AI จะวิเคราะห์จาก Current State แล้วให้คำแนะนำเชิงกลยุทธ์

### AI Behavior Rules
| Rule | Description |
|------|-------------|
| **Hint Style** | แนะนำ logic และ possible range เท่านั้น (**ห้ามบอกคำตอบตรงๆ**) |
| **Language** | **Multilingual** - ตอบกลับในภาษาเดียวกับที่ผู้เล่นใช้ |
| **Persona** | Friendly Game Master ที่คอยช่วยเหลือและให้กำลังใจ |

### Game Design
- ทำให้ "Chat" มี "Cooldown":
  - UI Cooldown: หลังจากกดส่งข้อความ ปุ่ม Chat จะ Disable เป็นเวลา 10-15 วินาที
  - Visual Cost: แสดงหลอด "Energy" หรือ "Mana" ของ AI ให้ผู้เล่นเห็นชัดๆ ว่าคุยได้อีกกี่ครั้ง (เช่น 1 เกม คุยได้ 10 ครั้ง)

### System Prompt (Phase 2)

```
You are a friendly Game Master for the Sum-100 puzzle game.

RULES:
- Respond in the SAME LANGUAGE the user writes in
- Give hints about logic and possible number ranges only
- NEVER reveal exact answers
- Be encouraging and supportive
- Keep responses concise

AVAILABLE FUNCTIONS:
- fill_cells: Fill numbers into grid cells
- verify_grid: Check all answers
- get_current_state: Get current grid state
```

---

## 🧩 5. Puzzle System

### Puzzle Source Strategy: **Hybrid Approach**
| Component | Description |
|-----------|-------------|
| **Puzzle Pool** | เตรียม 50-100 puzzles ที่ verified ว่ามี solution |
| **Selection** | Random เลือกจาก pool |
| **Variation** | Shuffle ตำแหน่ง pre-filled cells ให้ดูต่างกัน |

### Puzzle Generation Rules
- ทุก puzzle ต้องมี unique solution
- Pre-filled 9 ช่อง กระจายอย่างน้อย 1-2 ช่องต่อแถว/คอลัมน์
- ตัวเลขทั้งหมดอยู่ในช่วง 1-99

---

## 🛠️ 6. Technical Stack

### Frontend
| Technology | Version | Purpose |
|------------|---------|---------|
| **PhaserJS** | 3.90.0 | Game Engine, Grid UI, Effects |
| **React** | 18.x | UI Components |
| **TypeScript** | 5.x | Type Safety |
| **Tailwind CSS** | 3.x | Styling |
| **Apollo Client** | 3.x | GraphQL Client |
| **Zustand** | 4.x | Local State Management |
| **Vite** | 5.x | Build Tool |
| **WebLLM** | latest | Client-side AI (Phase 2) |

### Backend
| Technology | Version | Purpose |
|------------|---------|---------|
| **Golang** | 1.25.5 | API Server |
| **gqlgen** | latest | GraphQL Server |
| **Chi** | 5.x | HTTP Router |
| **pgx** | 5.x | PostgreSQL Driver |

### Database
| Technology | Purpose |
|------------|---------|
| **PostgreSQL** | Game State, Puzzle Pool, Leaderboard |

### DevOps
| Item | Choice |
|------|--------|
| **Containerization** | Docker Compose (local dev) |
| **Hosting** | Render (Backend + DB) + Vercel (Frontend) - Free Tier |
| **CI/CD** | ไม่ทำในเฟสแรก |

---

## 📱 7. Platform & UI

### Responsive Design
| Platform | Layout |
|----------|--------|
| **Desktop** | Side-by-side: Game (left) + Chat (right - Phase 2) |
| **Mobile** | Stacked: Game (top) + Chat (bottom sheet - Phase 2) |

### UI Language
| Component | Language |
|-----------|----------|
| **All UI Elements** | English only (buttons, labels, messages) |
| **AI Chat Responses** | Multilingual (based on user input - Phase 2) |

### Input Method
- **Desktop**: Click cell → Type number
- **Mobile**: Tap cell → Number pad popup

---

## 📊 8. GraphQL API (Updated)

### Queries
```graphql
type Query {
  """Get a game by ID"""
  game(id: ID!): Game

  """Get available puzzles for a guest with their status - Returns up to 20 puzzles ordered by ID"""
  getAvailablePuzzles(guestId: ID!, limit: Int = 20): [PuzzleWithStatus!]!

  """Get all puzzles from puzzle pool"""
  puzzles: [Puzzle!]!

  """Get leaderboard entries"""
  leaderboard(limit: Int = 10): [LeaderboardEntry!]!

  """Get statistics for a specific puzzle"""
  puzzleStats(puzzleId: ID!): PuzzleStats!

  """Get statistics for a specific player"""
  playerStats(guestId: ID!): PlayerStats!
}
```

### Mutations
```graphql
type Mutation {
  """Create a new game"""
  createGame(guestId: ID!): Game!

  """Fill cells with values"""
  fillCells(gameId: ID!, cells: [CellInput!]!): Game!

  """Make a move - fill a single cell and immediately verify it"""
  makeMove(gameId: ID!, row: Int!, col: Int!, value: Int!): MoveResult!

  """Verify all cells in the game"""
  verifyGame(gameId: ID!): VerifyResult!

  """Submit multiple answers for a puzzle - Main mutation for new flow"""
  submitAnswer(guestId: ID!, puzzleId: ID!, answers: [CellInput!]!): SubmitAnswerResult!

  """Complete the game and submit to leaderboard"""
  completeGame(gameId: ID!, guestId: ID!, username: String!): CompleteGameResult!

  """Unlock new puzzles after watching ad - Archives completed puzzles and unlocks AD_BLOCK puzzles"""
  unlockPuzzlesAfterAd(guestId: ID!): Boolean!
}
```

### Types
```graphql
"""Game represents a game session"""
type Game {
  id: ID!
  guestId: ID!
  puzzleId: ID!
  grid: [[Cell!]!]!
  totalMistakes: Int!
  status: GameStatus!
  createdAt: String!
}

"""Game result for public API (excludes GridSolution)"""
type GameResult {
  id: ID!
  guestId: ID!
  puzzleId: ID!
  gridCurrent: [[Cell!]!]!
  totalMistakes: Int!
  status: GameStatus!
  createdAt: String!
  updatedAt: String!
}

"""Result of submitting multiple answers"""
type SubmitAnswerResult {
  game: GameResult!
  results: [CellVerifyResult!]!
}

"""Cell represents a single cell in the grid"""
type Cell {
  row: Int!
  col: Int!
  value: Int
  isPreFilled: Boolean!
  feedback: CellFeedback
}

"""CellVerifyResult represents the verification result for a single cell"""
type CellVerifyResult {
  row: Int!
  col: Int!
  feedback: CellFeedback!
}

"""CellInput represents input for filling a cell"""
input CellInput {
  row: Int!
  col: Int!
  value: Int!
}

"""Game status enum"""
enum GameStatus {
  PLAYING
  COMPLETED
}

"""Cell feedback after verification"""
enum CellFeedback {
  CORRECT
  TOO_LOW
  TOO_HIGH
}

"""Puzzle with status for a guest"""
type PuzzleWithStatus {
  puzzle: Puzzle!
  status: PuzzleStatus!
}

"""Puzzle status enum"""
enum PuzzleStatus {
  AVAILABLE  # Not started yet
  PLAYING    # Currently playing
  COMPLETED  # Completed
  ARCHIVED   # Archived after ad unlock
  AD_BLOCK   # Available after watching ad
}
```

### New Game Flow (Recommended)
1. **Initial Load**:
   ```graphql
   query {
     getAvailablePuzzles(guestId: "xxx", limit: 20) {
       puzzle {
         id
         grid
         prefilledPositions
       }
       status  # AVAILABLE, PLAYING, COMPLETED, ARCHIVED, or AD_BLOCK
     }
   }
   ```

2. **Submit Answers**:
   ```graphql
   mutation {
     submitAnswer(
       guestId: "xxx"
       puzzleId: "puzzle-1"
       answers: [
         {row: 0, col: 0, value: 10},
         {row: 0, col: 1, value: 15}
       ]
     ) {
       game {
         id
         puzzleId
         gridCurrent
         totalMistakes
         status
       }
       results {
         row
         col
         feedback
       }
     }
   }
   ```

3. **Unlock Puzzles After Watching Ad** (When all 20 puzzles are completed):
   ```graphql
   mutation {
     unlockPuzzlesAfterAd(guestId: "xxx")
   }
   ```
   This will:
   - Mark all COMPLETED puzzles as ARCHIVED
   - Mark all AD_BLOCK puzzles as AVAILABLE
   - Next call to `getAvailablePuzzles` will return the next 20 puzzles

### Authentication & Security
- **Session-based**: Backend สร้างและจัดการ session cookies
- **Client-generated UUID**: เบราว์เซอร์สร้าง UUID สำหรับ initial session
- **Username**: ผู้เล่นระบุเมื่อจบเกมเพื่อบันทึกลง leaderboard
- **Solution Security**: ไม่ส่ง solution ไปยัง frontend - validate บน backend เท่านั้น
- **ไม่ใช้ IP, Fingerprint, หรือ tracking อื่นๆ**

---

## 🗄️ 9. Database Schema (Updated)

```sql
CREATE TABLE puzzle_pool (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  grid_solution JSONB NOT NULL,        -- 5x5 array of correct values
  prefilled_positions JSONB NOT NULL,  -- Array of {row, col} for pre-filled
  difficulty VARCHAR(20) DEFAULT 'medium',
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE game_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  puzzle_id UUID REFERENCES puzzle_pool(id),
  guest_id UUID NOT NULL,              -- UUID from client
  grid_current JSONB NOT NULL,         -- Current state of grid
  grid_solution JSONB NOT NULL,        -- Copy of solution
  prefilled_positions JSONB NOT NULL,
  total_mistakes INT DEFAULT 0,
  tokens_used INT DEFAULT 0,           -- AI tokens used (Phase 2)
  tokens_limit INT DEFAULT 1000,        -- AI token limit (Phase 2)
  status VARCHAR(20) DEFAULT 'playing', -- playing, completed
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE guest_puzzle_progress (
  guest_id UUID NOT NULL,
  puzzle_id UUID NOT NULL REFERENCES puzzle_pool(id),
  completed_at TIMESTAMP DEFAULT NOW(),
  status VARCHAR(20) NOT NULL DEFAULT 'COMPLETED',  -- COMPLETED, ARCHIVED, AD_BLOCK
  PRIMARY KEY (guest_id, puzzle_id)
);

CREATE TABLE leaderboard (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  game_session_id UUID REFERENCES game_sessions(id),
  guest_id UUID NOT NULL,              -- Add guest_id for tracking
  username VARCHAR(50) NOT NULL,
  mistakes INT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_leaderboard_mistakes ON leaderboard(mistakes ASC);
CREATE INDEX idx_game_sessions_guest_puzzle ON game_sessions(guest_id, puzzle_id);
```

### Table Descriptions

#### puzzle_pool
- Stores all available puzzles
- Contains solution grid and pre-filled positions
- Used to generate game sessions

#### game_sessions
- Tracks individual game sessions per guest per puzzle
- Each guest + puzzle combination can have multiple sessions
- Stores current grid state and progress
- Keeps copy of solution for validation
- Tracks total mistakes per session

#### guest_puzzle_progress
- Tracks which puzzles each guest has completed
- Tracks status: COMPLETED, ARCHIVED, AD_BLOCK
- Used by `getAvailablePuzzles` to show puzzles with their status
- Returns puzzles ordered by ID (not random) for consistent frontend display
- **COMPLETED**: Puzzle has been completed by the guest
- **ARCHIVED**: Previously completed, now archived after ad unlock
- **AD_BLOCK**: Available after watching ad (will become AVAILABLE after unlock)
- **AVAILABLE**: Not yet started (default for new puzzles)
- **PLAYING**: Currently being played by the guest

#### leaderboard
- Stores final results of completed games
- Sorted by total mistakes (fewer = better)
- Shows top 100 players globally

---

## 📁 10. Project Structure

```
/100sumgame
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── domain/          # Business entities
│   │   ├── repository/      # Data access
│   │   ├── service/         # Business logic
│   │   ├── graphql/         # Schema, Resolvers, Generated
│   │   ├── middleware/
│   │   └── database/
│   ├── pkg/
│   ├── gqlgen.yml
│   └── go.mod
│
├── frontend/
│   ├── src/
│   │   ├── game/            # PhaserJS scenes, objects, effects
│   │   ├── components/      # React UI components
│   │   ├── graphql/         # Queries, mutations, generated
│   │   ├── hooks/
│   │   ├── store/
│   │   ├── ai/              # WebLLM integration (Phase 2)
│   │   └── styles/
│   ├── codegen.ts
│   └── package.json
│
├── docker-compose.yml
└── README.md
```

---

## 🎯 11. Development Phases

### Phase 1: Core Game (Priority)
- ✅ Backend API (GraphQL, Database, Business Logic)
- ✅ Game UI (PhaserJS grid, Input handling)
- ✅ Game Mechanics (Fill cells, Verify, Complete)
- ✅ Leaderboard System
- ✅ Basic React Components
- ❌ AI Chat & WebLLM (Skip for Phase 1)

### Phase 2: AI Integration (After Phase 1 Complete)
- WebLLM setup in frontend
- Client-side function calling
- AI Chat interface
- Cooldown system
- Multilingual support
- Energy/Mana system

---

## 📝 12. Notes

- **AI ทำใน client-side**: ไม่มี dependency กับ backend AI services
- **UUID tracking**: Client สร้าง UUID และส่งมากับทุก request
- **Username optional**: ผู้เล่นสามารถเล่นโดยไม่ต้องระบุ username จนกว่าจะจบเกม
- **No tracking**: ไม่มีการเก็บข้อมูล IP, fingerprint, หรือ tracking อื่นๆ