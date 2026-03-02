/*

POST /comments – создание комментария
{
    "parent_id": 0,
    "text": "Текст комментария"
}

GET /comments?parent={id} – получение комментария и вложенных

DELETE /comments/{id} – удаление комментария и всех вложенных

GET /search?q=текст

*/

package api
