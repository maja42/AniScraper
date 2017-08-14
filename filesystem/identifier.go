package filesystem

import (
	"context"
	"fmt"
	"sync"
)

type AnimeIdentifier interface {
	Output() chan<- *AnimeFolder
	Start(ctx context.Context, routineCount int) error
}

type animeIdentifier struct {
	mutex   sync.RWMutex
	started bool
	ctx     context.Context

	webClient WebClient

	routineWg sync.WaitGroup // counts all currently running identifier go routines
	tasks     <-chan *AnimeFolder
	output    chan *AnimeFolder
}

func NewAnimeIdentifier(webClient WebClient, tasks <-chan *AnimeFolder) AnimeIdentifier {
	return &animeIdentifier{
		webClient: webClient,
	}
}

func (ident *animeIdentifier) Output() chan<- *AnimeFolder {
	return ident.output
}

func (ident *animeIdentifier) Start(ctx context.Context, routineCount int) error {
	ident.mutex.Lock()
	defer ident.mutex.Unlock()

	if ident.started {
		return fmt.Errorf("The AnimeIdentifier already started")
	}
	if routineCount <= 0 {
		return fmt.Errorf("Invalid routine count")
	}

	ident.started = true
	ident.ctx = ctx

	ident.output = make(chan *AnimeFolder, 20)

	for i := 0; i < routineCount; i++ {
		ident.routineWg.Add(1)
		go ident.identify(i)
	}

	go func() {
		ident.routineWg.Wait()
		close(ident.output)
	}()

	return nil
}

func (ident *animeIdentifier) identify(identifierId int) {
	defer ident.routineWg.Done()

}

// func (webcli *HttpWebClient) IdentifyAnime(needle string) ([]AnimeSearchResult, error) {
//     const itemSelector = "li[itemscope]" // Returns one element for each search result
//     const titleSel = "th a,div"
//     const animeDetailsSel = "td[data-title=\"Typ / Episoden / Jahr\"]"
//     const rankSel = "td[data-title=\"Rang\"]"

//     url := webcli.urlProvider.GetUrl_Search(needle)
//     doc, err := webcli.getDocument(url)
//     if err != nil {
//         return nil, err
//     }

//     findings := make([]AnimeSearchResult, 0)

//     // Analyse page
//     doc.Find(itemSelector).Each(func(i int, s *goquery.Selection) {
//         result := NewAnimeSearchResult()
//         var num int

//         // Fetch title
//         node := s.Find("h2")
//         result.Title = strings.TrimSpace(OwnTextOnly(node))
//         if err := IsValidTitle(result.Title); err != nil {
//             log.Warningf("Could not find any title for search result (needle: '%v'): %v", needle, err)
//             return
//         }

//         // Fetch Url
//         var exists bool
//         node = s.Find(".title").Parent()
//         result.HRef, exists = node.Attr("href")
//         if !exists {
//             log.Warningf("Could not get anime url for search result (title: '%v') - ignoring", result.Title)
//             return
//         }

//         // Fetch details
//         text := node.Children().Text()
//         groups := webcli.searchDetailsRegex.FindStringSubmatch(text)
//         if len(groups) != 4 {
//             log.Warningf("Could not get type, episode count and year of search result (title: '%v')", result.Title)
//         } else {
//             //type
//             result.AnimeType = ConvertToAnimeType(groups[1])
//             //episodes
//             if num, err = strconv.Atoi(groups[2]); err != nil {
//                 log.Warningf("Could not find episode count of search result (title: '%v'): %v", result.Title, err)
//             } else if err := IsValidEpisodeCount(num); err != nil {
//                 log.Warningf("Could not set episode count of search result (title: '%v'): %v", result.Title, err)
//             } else {
//                 result.Episodes = num
//             }
//             // year
//             if num, err = strconv.Atoi(groups[3]); err != nil {
//                 log.Warningf("Could not find year of search result (title: '%v'): %v", result.Title, err)
//             } else if err := IsValidYear(num); err != nil {
//                 log.Warningf("Could not set year of search result (title: '%v'): %v", result.Title, err)
//             } else {
//                 result.Year = num
//             }
//         }

//         // Fetch description
//         node = s.Find(".text")
//         text = strings.TrimSpace(OwnTextOnly(node))
//         if err := IsValidDescription(text); err != nil {
//             if !(node.Length() == 1 && text == "") {
//                 log.Warningf("Could not find description for search result (title: '%v'): %v", result.Title, err)
//             }
//         } else {
//             result.Description = text
//         }

//         // Fetch duration
//         node = s.Find(".episodes")
//         text = strings.TrimSpace(OwnTextOnly(node))
//         durationInMins, err := GetDurationInMinsFromString(text)
//         if err != nil {
//             log.Warningf("Could not find duration for search result (title: '%v'): %v", result.Title, err)
//         } else if err := IsValidDuration(durationInMins); err != nil {
//             log.Warningf("Could not set duration of search result (title: '%v'): %v", result.Title, err)
//         } else {
//             result.DurationInMins = durationInMins
//         }

//         // Fetch creator
//         node = s.Find("[itemprop=creator] [itemprop=name]")
//         text = strings.TrimSpace(OwnTextOnly(node))
//         if len(text) > 0 {
//             if err := IsValidCreator(text); err != nil {
//                 log.Warningf("Could not find creator for search result (title: '%v'): %v", result.Title, err)
//             } else {
//                 result.Creator = text
//             }
//         }

//         findings = append(findings, *result)
//     })
//     return findings, nil
// }
