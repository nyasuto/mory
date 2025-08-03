"""Summarization service for automatic text summarization using OpenAI API"""

import asyncio
from typing import Any

import openai

from ..core.config import settings
from ..models.memory import Memory


class SummarizationService:
    """Service for generating text summaries using OpenAI API"""

    def __init__(self):
        """Initialize summarization service"""
        self.call_count = 0
        self.enabled = getattr(settings, "summary_enabled", True)
        self.model = getattr(settings, "summary_model", "gpt-4-turbo")
        self.max_length = getattr(settings, "summary_max_length", 200)
        self.fallback_enabled = getattr(settings, "summary_fallback_enabled", True)

        if self.enabled and hasattr(settings, "openai_api_key"):
            openai.api_key = settings.openai_api_key

    async def generate_summary(
        self, text: str, max_length: int | None = None, language: str = "ja"
    ) -> str:
        """
        Generate a summary of the given text

        Args:
            text: Text to summarize
            max_length: Maximum length of summary (defaults to config)
            language: Language for summary ("ja" or "en")

        Returns:
            Generated summary text

        Raises:
            Exception: If OpenAI API fails and fallback is disabled

        """
        if not self.enabled:
            return self._fallback_summary(text, max_length)

        max_len = max_length or self.max_length

        # If text is already short enough, return with prefix
        if len(text) <= max_len:
            prefix = "要約: " if language == "ja" else "Summary: "
            return f"{prefix}{text}"

        try:
            self.call_count += 1

            # Create prompt based on language
            prompt = self._create_prompt(text, max_len, language)

            # Call OpenAI API
            response = await self._call_openai_api(prompt)

            # Extract and validate summary
            summary = self._extract_summary(response, max_len)

            return summary

        except Exception as e:
            if self.fallback_enabled:
                return self._fallback_summary(text, max_len, language)
            else:
                raise Exception(f"Summary generation failed: {str(e)}") from e

    async def should_regenerate_summary(self, memory: Memory) -> bool:
        """
        Check if summary should be regenerated

        Args:
            memory: Memory object to check

        Returns:
            True if summary should be regenerated

        """
        # Regenerate if no summary exists
        if not memory.summary:
            return True

        # Regenerate if summary is significantly shorter than expected
        if len(memory.summary) < 20:
            return True

        # Regenerate if memory was updated after summary generation
        if memory.summary_generated_at and memory.updated_at > memory.summary_generated_at:
            return True

        return False

    async def batch_generate_summaries(
        self, memories: list[Memory], delay_ms: float = 100
    ) -> dict[str, str]:
        """
        Generate summaries for multiple memories with rate limiting

        Args:
            memories: List of Memory objects
            delay_ms: Delay between requests in milliseconds

        Returns:
            Dictionary mapping memory IDs to summaries

        """
        results = {}

        for memory in memories:
            try:
                summary = await self.generate_summary(str(memory.value))
                results[str(memory.id)] = summary

                # Rate limiting delay
                if delay_ms > 0:
                    await asyncio.sleep(delay_ms / 1000.0)

            except Exception as e:
                results[str(memory.id)] = f"Error: {str(e)}"

        return results

    def _create_prompt(self, text: str, max_length: int, language: str) -> str:
        """Create prompt for OpenAI API based on language"""
        prompts = {
            "ja": f"""以下のテキストを{max_length}文字程度で要約してください。
重要なポイントを押さえ、簡潔で分かりやすい日本語で表現してください。
要約は「要約: 」で始めてください。

テキスト: {text}

要約:""",
            "en": f"""Please summarize the following text in approximately {max_length} characters.
Focus on key points and express in clear, concise English.
Start the summary with "Summary: ".

Text: {text}

Summary:""",
        }

        return prompts.get(language, prompts["ja"])

    async def _call_openai_api(self, prompt: str) -> str:
        """Call OpenAI Chat Completion API"""
        try:
            # Use the new OpenAI client API
            client = openai.AsyncOpenAI(api_key=settings.openai_api_key)

            response = await client.chat.completions.create(
                model=self.model,
                messages=[{"role": "user", "content": prompt}],
                max_tokens=100,  # Limit response tokens for summaries
                temperature=0.3,  # Lower temperature for consistent summaries
            )

            content = response.choices[0].message.content
            return content.strip() if content else ""

        except Exception as e:
            raise Exception(f"OpenAI API call failed: {str(e)}") from e

    def _extract_summary(self, response: str, max_length: int) -> str:
        """Extract and validate summary from API response"""
        if not response:
            raise Exception("Empty response from OpenAI API")

        # Clean up the response
        summary = response.strip()

        # Ensure summary doesn't exceed max length
        if len(summary) > max_length:
            summary = summary[: max_length - 3] + "..."

        # Ensure Japanese summaries have proper prefix
        if "要約:" not in summary and any(ord(c) > 127 for c in summary):
            if not summary.startswith("要約: "):
                summary = f"要約: {summary}"

        return summary

    def _fallback_summary(
        self, text: str, max_length: int | None = None, language: str = "ja"
    ) -> str:
        """Generate fallback summary when API is unavailable"""
        max_len = max_length or self.max_length

        if len(text) <= max_len:
            prefix = "要約: " if language == "ja" else "Summary: "
            return f"{prefix}{text}"

        # Simple truncation with ellipsis
        truncated = text[: max_len - 10]

        # Try to cut at word boundary
        if " " in truncated:
            last_space = truncated.rfind(" ")
            if last_space > max_len * 0.8:  # Only if we don't lose too much
                truncated = truncated[:last_space]

        prefix = "要約: " if language == "ja" else "Summary: "
        return f"{prefix}{truncated}..."

    def get_stats(self) -> dict[str, Any]:
        """Get service statistics"""
        return {
            "enabled": self.enabled,
            "model": self.model,
            "max_length": self.max_length,
            "fallback_enabled": self.fallback_enabled,
            "call_count": self.call_count,
        }


# Global service instance
summarization_service = SummarizationService()
