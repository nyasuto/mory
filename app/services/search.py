"""Search service for memory search functionality"""

import time

import numpy as np
import openai
from sqlalchemy import and_, or_, text
from sqlalchemy.orm import Session

from ..core.config import settings
from ..core.database import check_fts5_support
from ..models.memory import Memory
from ..models.schemas import MemoryResponse, SearchRequest, SearchResponse, SearchResult


class SearchService:
    """Service for memory search operations"""

    def __init__(self) -> None:
        """Initialize search service with available search backends"""
        self.fts5_available = check_fts5_support()
        self.semantic_available = settings.is_semantic_available
        if self.semantic_available:
            openai.api_key = settings.openai_api_key

    async def search_memories(self, request: SearchRequest, db: Session) -> SearchResponse:
        """Perform memory search with specified type"""
        start_time = time.time()

        # Determine search strategy
        search_type = self._determine_search_type(request.search_type)

        results: list[SearchResult] = []
        total = 0

        if search_type == "fts5":
            results, total = await self._search_fts5(request, db)
        elif search_type == "semantic":
            results, total = await self._search_semantic(request, db)
        elif search_type == "hybrid":
            results, total = await self._search_hybrid(request, db)
        else:
            # Fallback to LIKE search
            results, total = await self._search_like(request, db)

        execution_time = (time.time() - start_time) * 1000

        return SearchResponse(
            results=results,
            total=total,
            query=request.query,
            search_type=search_type,
            execution_time_ms=round(execution_time, 2),
            filters={
                "tags": request.tags,
                "date_from": request.date_from.isoformat() if request.date_from else None,
                "date_to": request.date_to.isoformat() if request.date_to else None,
            },
        )

    def _determine_search_type(self, requested_type: str) -> str:
        """Determine the actual search type to use"""
        if requested_type == "semantic" and not self.semantic_available:
            return "fts5" if self.fts5_available else "like"
        elif requested_type == "fts5" and not self.fts5_available:
            return "like"
        elif requested_type == "hybrid" and not (self.fts5_available or self.semantic_available):
            return "like"
        return requested_type

    async def _search_fts5(
        self, request: SearchRequest, db: Session
    ) -> tuple[list[SearchResult], int]:
        """Perform FTS5 full-text search"""
        if not self.fts5_available:
            return await self._search_like(request, db)

        # Build FTS5 query
        fts_query = self._build_fts5_query(request.query)

        # Build filter conditions and parameters
        filter_conditions, filter_params = self._build_fts5_filters(request)

        # Build the main query
        base_sql = """
            SELECT m.*, fts.rank
            FROM memories m
            JOIN memories_fts fts ON m.id = fts.id
            WHERE memories_fts MATCH :query
        """

        if filter_conditions:
            query = text(f"{base_sql} AND {filter_conditions}")
        else:
            query = text(base_sql)

        # Prepare parameters
        params = {"query": fts_query}
        params.update(filter_params)

        # Execute search
        result = db.execute(query, params)
        rows = result.fetchall()

        # Convert to SearchResult objects
        results = []
        for row in rows:
            memory = Memory()
            for key, value in row._mapping.items():
                if hasattr(memory, key) and key != "rank":
                    setattr(memory, key, value)

            results.append(
                SearchResult(
                    memory=MemoryResponse.model_validate(memory),
                    score=max(0.1, min(abs(float(row.rank)) / 10.0, 1.0)),  # Normalize FTS5 rank
                    search_type="fts5",
                )
            )

        # Apply pagination
        total = len(results)
        paginated_results = results[request.offset : request.offset + request.limit]

        return paginated_results, total

    async def _search_semantic(
        self, request: SearchRequest, db: Session
    ) -> tuple[list[SearchResult], int]:
        """Perform semantic search using OpenAI embeddings"""
        if not self.semantic_available:
            return await self._search_fts5(request, db)

        try:
            # Generate embedding for query
            response = openai.embeddings.create(model=settings.openai_model, input=request.query)
            query_embedding = response.data[0].embedding

            # Get memories with embeddings
            query = db.query(Memory).filter(Memory.embedding.isnot(None))

            # Apply filters
            query = self._apply_filters(query, request)

            memories = query.all()

            # Calculate similarities
            results = []
            for memory in memories:
                if memory.embedding:
                    memory_embedding = np.frombuffer(memory.embedding, dtype=np.float32)
                    similarity = self._cosine_similarity(query_embedding, memory_embedding)

                    if similarity > 0.1:  # Minimum similarity threshold
                        results.append(
                            SearchResult(
                                memory=MemoryResponse.model_validate(memory),
                                score=float(similarity),
                                search_type="semantic",
                            )
                        )

            # Sort by similarity
            results.sort(key=lambda x: x.score, reverse=True)

            # Apply pagination
            total = len(results)
            paginated_results = results[request.offset : request.offset + request.limit]

            return paginated_results, total

        except Exception as e:
            print(f"Semantic search failed: {e}")
            return await self._search_fts5(request, db)

    async def _search_hybrid(
        self, request: SearchRequest, db: Session
    ) -> tuple[list[SearchResult], int]:
        """Perform hybrid search combining FTS5 and semantic search"""
        # Get results from both search types
        fts_results, _ = await self._search_fts5(request, db)
        semantic_results, _ = await self._search_semantic(request, db)

        # Combine and re-rank results
        combined_results = {}

        # Add FTS5 results with weight
        for result in fts_results:
            memory_id = result.memory.id
            combined_results[memory_id] = SearchResult(
                memory=result.memory,
                score=result.score * 0.3,  # 30% weight for FTS5
                search_type="hybrid",
            )

        # Add semantic results with weight
        for result in semantic_results:
            memory_id = result.memory.id
            if memory_id in combined_results:
                # Combine scores
                combined_results[memory_id].score += result.score * 0.7  # 70% weight for semantic
            else:
                combined_results[memory_id] = SearchResult(
                    memory=result.memory, score=result.score * 0.7, search_type="hybrid"
                )

        # Sort by combined score
        results = list(combined_results.values())
        results.sort(key=lambda x: x.score, reverse=True)

        # Apply pagination
        total = len(results)
        paginated_results = results[request.offset : request.offset + request.limit]

        return paginated_results, total

    async def _search_like(
        self, request: SearchRequest, db: Session
    ) -> tuple[list[SearchResult], int]:
        """Fallback LIKE search when FTS5 is not available"""
        query = db.query(Memory)

        # Build LIKE conditions
        search_terms = request.query.split()
        like_conditions = []

        for term in search_terms:
            like_pattern = f"%{term}%"
            like_conditions.append(
                or_(
                    Memory.value.ilike(like_pattern),
                    Memory.summary.ilike(like_pattern),
                    Memory.tags.ilike(like_pattern),
                )
            )

        if like_conditions:
            query = query.filter(and_(*like_conditions))

        # Apply other filters
        query = self._apply_filters(query, request)

        # Get total count
        total = query.count()

        # Apply pagination and ordering
        memories = (
            query.order_by(Memory.updated_at.desc())
            .offset(request.offset)
            .limit(request.limit)
            .all()
        )

        # Convert to SearchResult objects
        results = []
        for memory in memories:
            # Simple relevance scoring based on term frequency
            score = self._calculate_like_score(memory, search_terms)
            results.append(
                SearchResult(
                    memory=MemoryResponse.model_validate(memory), score=score, search_type="like"
                )
            )

        return results, total

    def _build_fts5_query(self, query: str) -> str:
        """Build FTS5 query string"""
        # Split query into terms and escape special characters
        terms = query.split()
        escaped_terms = []

        for term in terms:
            # Remove special FTS5 characters and quote terms
            escaped_term = term.replace('"', "").replace("'", "")
            if escaped_term:
                escaped_terms.append(f'"{escaped_term}"')

        return " ".join(escaped_terms)

    def _build_fts5_filters(self, request: SearchRequest) -> tuple[str, dict]:
        """Build parameterized WHERE clause filters for FTS5 query"""
        filters = []
        params = {}

        # Category filtering removed in simplified schema (Issue #112)

        if request.tags:
            tag_conditions = []
            for i, tag in enumerate(request.tags):
                param_name = f"tag_{i}"
                tag_conditions.append(f"m.tags LIKE :{param_name}")
                params[param_name] = f'%"{tag}"%'
            filters.append(f"({' OR '.join(tag_conditions)})")

        if request.date_from:
            filters.append("m.created_at >= :date_from")
            params["date_from"] = request.date_from.isoformat()

        if request.date_to:
            filters.append("m.created_at <= :date_to")
            params["date_to"] = request.date_to.isoformat()

        filter_sql = " AND ".join(filters) if filters else ""
        return filter_sql, params

    def _build_filters(self, request: SearchRequest) -> str:
        """Build WHERE clause filters for SQL query (legacy method for non-FTS5)"""
        filters = []

        # Category filtering removed in simplified schema (Issue #112)

        if request.tags:
            tag_conditions = []
            for tag in request.tags:
                tag_conditions.append(f"m.tags LIKE '%\"{tag}\"%'")
            filters.append(f"({' OR '.join(tag_conditions)})")

        if request.date_from:
            filters.append(f"m.created_at >= '{request.date_from.isoformat()}'")

        if request.date_to:
            filters.append(f"m.created_at <= '{request.date_to.isoformat()}'")

        return " AND ".join(filters) if filters else ""

    def _apply_filters(self, query, request: SearchRequest):
        """Apply filters to SQLAlchemy query"""
        # Category filtering removed in simplified schema (Issue #112)

        if request.tags:
            tag_conditions = []
            for tag in request.tags:
                tag_conditions.append(Memory.tags.ilike(f'%"{tag}"%'))
            query = query.filter(or_(*tag_conditions))

        if request.date_from:
            query = query.filter(Memory.created_at >= request.date_from)

        if request.date_to:
            query = query.filter(Memory.created_at <= request.date_to)

        return query

    def _cosine_similarity(self, a: list[float], b: np.ndarray) -> float:
        """Calculate cosine similarity between two vectors"""
        a_array = np.array(a, dtype=np.float32)
        return float(np.dot(a_array, b) / (np.linalg.norm(a_array) * np.linalg.norm(b)))

    def _calculate_like_score(self, memory: Memory, search_terms: list[str]) -> float:
        """Calculate relevance score for LIKE search"""
        content = f"{memory.value} {memory.category} {memory.key or ''} {memory.tags}"
        content_lower = content.lower()

        score = 0.0
        for term in search_terms:
            term_lower = term.lower()
            count = content_lower.count(term_lower)
            score += count * 0.1

        return min(score, 1.0)


# Global search service instance
search_service = SearchService()
